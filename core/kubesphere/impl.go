package kub

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	defaultPassword = "DefPwd123@"
)

// GetURL returns the Kubernetes URL.
func (m *Mgr) GetURL() string {
	return m.kubURL
}

// GetCluster returns the cluster.
func (m *Mgr) GetCluster() string {
	return m.curCluster
}

// CreateUserAccount creates a new user account with the provided details.
func (m *Mgr) CreateUserAccount(userAccount, password string) error {
	email := userAccount + "@titan.com"

	body := map[string]interface{}{
		"apiVersion": "iam.kubesphere.io/v1beta1",
		"kind":       "User",
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"iam.kubesphere.io/uninitialized": "true", "iam.kubesphere.io/globalrole": "platform-regular", "kubesphere.io/creator": "admin",
			}, "name": userAccount,
		},
		"spec": map[string]interface{}{"email": email, "password": password},
	}

	_, err := m.doRequest("POST", "/kapis/iam.kubesphere.io/v1beta1/users", body)
	if err != nil {
		log.Errorf("CreateUserAccount err:%s", err.Error())
		return err
	}

	// log.Infoln("CreateUserAccount rsp-----")
	// log.Infoln(string(rsp))

	return nil
}

// CreateSpaceAndResourceQuotas creates a space and resource quotas for a user.
func (m *Mgr) CreateSpaceAndResourceQuotas(workspaceID, userAccount, cluster string, cpu, ram, storage int) error {
	err := m.createUserSpace(workspaceID, userAccount, cluster)
	if err != nil {
		log.Errorf("CreateUserSpace: %s", err.Error())
		return err
	}

	time.Sleep(1 * time.Second)
	err = m.changeWorkspaceMembers(workspaceID, userAccount)
	if err != nil {
		log.Errorf("changeWorkspaceMembers: %s", err.Error())
		return err
	}

	err = m.createUserResourceQuotas(workspaceID, cluster, cpu, ram, storage)
	if err != nil {
		log.Errorf("CreateUserResourceQuotas: %s", err.Error())
	}

	return err
}

// createUserSpace creates a user space for the given workspaceID and user.
func (m *Mgr) createUserSpace(workspaceID, userAccount, cluster string) error {
	body := map[string]interface{}{
		"apiVersion": "iam.kubesphere.io/v1beta1",
		"kind":       "WorkspaceTemplate",
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"kubesphere.io/creator": "admin",
			}, "name": workspaceID,
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"manager": userAccount,
				},
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"kubesphere.io/creator": "admin",
					},
				},
			},
			"placement": map[string]interface{}{
				"clusters": []map[string]interface{}{
					{
						"name": cluster,
					},
				},
			},
		},
	}

	_, err := m.doRequest("POST", "/kapis/tenant.kubesphere.io/v1beta1/workspacetemplates", body)
	if err != nil {
		log.Errorf("CreateUserSpace err:%s", err.Error())
		return err
	}

	// log.Infoln("CreateUserSpace rsp-----")
	// log.Infoln(string(rsp))

	return nil
}

func (m *Mgr) changeWorkspaceMembers(workspaceID, userAccount string) error {
	body := map[string]interface{}{
		"roleRef":  fmt.Sprintf("%s-self-provisioner", workspaceID),
		"username": userAccount,
	}

	path := fmt.Sprintf("/kapis/iam.kubesphere.io/v1beta1/workspaces/%s/workspacemembers/%s", workspaceID, userAccount)
	_, err := m.doRequest("PUT", path, body)
	if err != nil {
		log.Errorf("changeWorkspaceMembers err:%s", err.Error())
		return err
	}

	// log.Infoln("changeWorkspaceMembers rsp-----")
	// log.Infoln(string(rsp))

	return nil
}

// createUserResourceQuotas creates resource quotas for a user.
// It takes an workspaceID string and resource limits for CPU, ram, and storage.
func (m *Mgr) createUserResourceQuotas(workspaceID, cluster string, cpu, ram, storage int) error {
	body := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				"kubesphere.io/workspace": workspaceID,
			}, "name": workspaceID,
		},
		"spec": map[string]interface{}{
			"selector": map[string]interface{}{
				"kubesphere.io/workspace": workspaceID,
			},
			"quota": map[string]interface{}{
				"hard": map[string]interface{}{
					"limits.cpu":             fmt.Sprintf("%d", cpu),
					"limits.memory":          fmt.Sprintf("%dGi", ram),
					"requests.cpu":           fmt.Sprintf("%d", cpu),
					"requests.memory":        fmt.Sprintf("%dGi", ram),
					"requests.storage":       fmt.Sprintf("%dGi", storage),
					"persistentvolumeclaims": fmt.Sprintf("%d", storage),
				},
			},
		},
	}

	path := fmt.Sprintf("/clusters/%s/kapis/tenant.kubesphere.io/v1beta1/workspaces/%s/resourcequotas", cluster, workspaceID)
	_, err := m.doRequest("POST", path, body)
	if err != nil {
		log.Errorf("CreateUserResourceQuotas err:%s", err.Error())
		return err
	}

	// log.Infoln("CreateUserResourceQuotas rsp-----")
	// log.Infoln(string(rsp))

	return nil
}

// DeleteUserSpace removes a user space based on the provided workspaceID.
func (m *Mgr) DeleteUserSpace(workspaceID, cluster string) error {
	body := map[string]interface{}{
		"apiVersion":        "iam.kubesphere.io/v1beta1",
		"kind":              "DeleteOptions",
		"propagationPolicy": "Orphan",
	}

	path := fmt.Sprintf("/kapis/tenant.kubesphere.io/v1beta1/workspacetemplates/%s", workspaceID)
	_, err := m.doRequest("DELETE", path, body)
	if err != nil {
		log.Errorf("DeleteUserSpace err:%s", err.Error())
		// return err
	}

	err = m.deleteProjects(workspaceID, cluster)
	if err != nil {
		log.Errorf("DeleteUserSpace deleteProjects err:%s", err.Error())
	}

	// log.Infoln("DeleteUserSpace rsp-----")
	// log.Infoln(string(rsp))

	return nil
}

// ResetPassword resets the password for the specified user.
func (m *Mgr) ResetPassword(userAccount, password string) error {
	body := map[string]interface{}{
		"password":   password,
		"rePassword": password,
	}

	path := fmt.Sprintf("/kapis/iam.kubesphere.io/v1beta1/users/%s/password", userAccount)
	rsp, err := m.doRequest("PUT", path, body)
	if err != nil {
		log.Errorf("ResetPassword err:%s", err.Error())
		return err
	}

	log.Infoln("ResetPassword rsp-----")
	log.Infoln(string(rsp))

	return nil
}

type resourceRsp struct {
	Metadata struct {
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
}

func (m *Mgr) getUserResourceQuotas(workspaceID, cluster string) (string, error) {
	path := fmt.Sprintf("/clusters/%s/kapis/tenant.kubesphere.io/v1beta1/workspaces/%s/resourcequotas/%s", cluster, workspaceID, workspaceID)
	rsp, err := m.doRequest("GET", path, nil)
	if err != nil {
		log.Errorf("getUserResourceQuotas err:%s", err.Error())
		return "", err
	}

	var info resourceRsp
	err = json.Unmarshal(rsp, &info)
	if err != nil {
		return "", err
	}

	// log.Infoln("getUserResourceQuotas rsp-----")
	// log.Infoln(string(rsp))

	return info.Metadata.ResourceVersion, nil
}

// UpdateUserResourceQuotas update resource quotas for a user.
// It takes an workspaceID string and resource limits for CPU, ram, and storage.
func (m *Mgr) UpdateUserResourceQuotas(workspaceID, cluster string, cpu, ram, storage int) error {
	ver, err := m.getUserResourceQuotas(workspaceID, cluster)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"apiVersion": "quota.kubesphere.io/v1alpha2",
		"kind":       "ResourceQuota",
		"metadata": map[string]interface{}{
			"name":            workspaceID,
			"cluster":         cluster,
			"resourceVersion": ver,
			"workspace":       workspaceID,
		},
		"spec": map[string]interface{}{
			"quota": map[string]interface{}{
				"hard": map[string]interface{}{
					"limits.cpu":             fmt.Sprintf("%d", cpu),
					"limits.memory":          fmt.Sprintf("%dGi", ram),
					"requests.cpu":           fmt.Sprintf("%d", cpu),
					"requests.memory":        fmt.Sprintf("%dGi", ram),
					"requests.storage":       fmt.Sprintf("%dGi", storage),
					"persistentvolumeclaims": fmt.Sprintf("%d", storage),
				},
			},
		},
	}

	path := fmt.Sprintf("/clusters/%s/kapis/tenant.kubesphere.io/v1beta1/workspaces/%s/resourcequotas/%s", cluster, workspaceID, workspaceID)
	_, err = m.doRequest("PUT", path, body)
	if err != nil {
		log.Errorf("UpdateUserResourceQuotas err:%s", err.Error())
		return err
	}

	// log.Infoln("UpdateUserResourceQuotas rsp-----")
	// log.Infoln(string(rsp))

	return nil
}

type projectRsp struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
	} `json:"items"`
	TotalItems int `json:"totalItems"`
}

func (m *Mgr) listProjects(workspaceID, cluster string) (*projectRsp, error) {
	path := fmt.Sprintf("/clusters/%s/kapis/tenant.kubesphere.io/v1beta1/workspaces/%s/namespaces", cluster, workspaceID)
	rsp, err := m.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	// log.Infoln("listProjects rsp-----")
	// log.Infoln(string(rsp))

	var info projectRsp
	err = json.Unmarshal(rsp, &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// deleteProjects removes projects from the specified workspace and cluster.
func (m *Mgr) deleteProjects(workspaceID, cluster string) error {
	rsp, err := m.listProjects(workspaceID, cluster)
	if err != nil {
		return err
	}

	for _, info := range rsp.Items {
		path := fmt.Sprintf("/clusters/%s/kapis/tenant.kubesphere.io/v1beta1/workspaces/%s/namespaces/%s", cluster, workspaceID, info.Metadata.Name)
		_, err := m.doRequest("DELETE", path, nil)
		if err != nil {
			log.Errorf("DeleteProjects workspaceID:[%s],name:[%s] err:%s", workspaceID, info.Metadata.Name, err.Error())
			continue
		}

		// log.Infoln("DeleteProjects rsp-----")
		// log.Infoln(string(rsp))
	}

	return nil
}
