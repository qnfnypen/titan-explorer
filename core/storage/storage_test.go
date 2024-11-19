package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/pkg/random"
	"github.com/google/uuid"
)

func TestAesEncryptCBC(t *testing.T) {
	uid := "leeyfann@gmail.com"
	ui := UserKeyInfo{UID: uid, Salt: random.GenerateRandomString(6)}
	uk, _ := json.Marshal(ui)
	ak, err := AesEncryptCBC(uk, cryptoKey)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(ak)

	nuid, err := AesDecryptCBCByKey(ak)
	if err != nil {
		t.Fatal(err)
	}

	if nuid != uid {
		t.Fail()
	}
}

func TestGet(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://storage-test.titannet.io/api/v1/country_count", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("apiKey", "1W6gH1BfRGNrgt4REZV7VpucVp3gcbkkSs39jC4+m9ejNVmKKJHxMJ+AIy/Q8mI5")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp.StatusCode)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(body))
}

func TestCIDToHash(t *testing.T) {
	cid := "bafkreib5arnexhnsn6etb4xs7ywm52iey3i7xkxgjxm4dhw5njxmz2dn4i"

	hash, err := CIDToHash(cid)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hash)
}

func TestHashToCID(t *testing.T) {
	hash := "1220596d64b362871d9b1b748f0044ffb5ef0e54df29271268c2875459d80c71e8be"

	cid, err := HashToCID(hash)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(cid)
}

func TestAesDecryptCBCByKey(t *testing.T) {
	uid, err := AesDecryptCBCByKey("uAUvW5pb6CM0/dgozgxJNChc5xQYNCsNT9RK4csTDvevPv1fHyLX5C9QLw9jQnK/WZMqRJWtTp+yxJ1/x/gEyvIdtPYQ9lpAQsW8UQhOE5w=")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(uid)
}

func TestCreateTenant(t *testing.T) {
	tenant := &model.Tenant{
		TenantID:        uuid.NewString(),
		Name:            "tontdrive",
		State:           dao.TenantStateActive,
		UploadNotifyUrl: "https://dev.t-drive.io/api/v1/bot/titanCallback",
		CreatedAt:       time.Now(),
		// ApiKey:          uuid.NewString(),
	}

	apiKeySecretPairBlob, key, secret, err := CreateTenantKey(tenant.TenantID, tenant.Name)

	fmt.Println(apiKeySecretPairBlob, key, secret, err)

	sql, args, err := squirrel.Insert("tenants").Columns("tenant_id, name, api_key, state, upload_notify_url, created_at").Values(
		tenant.TenantID, tenant.Name, apiKeySecretPairBlob, tenant.State, tenant.UploadNotifyUrl, tenant.CreatedAt,
	).ToSql()

	fmt.Println(sql, args, err)

	fmt.Printf("%x\n", apiKeySecretPairBlob)
}
