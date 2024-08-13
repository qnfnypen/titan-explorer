package opfie

import (
	"context"
	"fmt"
	"strings"

	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/kubo/client/rpc"
	ma "github.com/multiformats/go-multiaddr"
)

// IPFSClient ipfs客户端
type IPFSClient struct {
	node *rpc.HttpApi
}

// NewIPFSClient 新建ipfs客户端
func NewIPFSClient(url string) (*IPFSClient, error) {
	a, err := ma.NewMultiaddr(strings.TrimSpace(url))
	if err != nil {
		return nil, fmt.Errorf("new multiaddr error:%w", err)
	}
	node, err := rpc.NewApi(a)
	if err != nil {
		return nil, fmt.Errorf("new ipfs node error:%w", err)
	}

	return &IPFSClient{node: node}, nil
}

// AddFileByCID 通过cid上传文件到节点
func (c *IPFSClient) AddFileByCID(ctx context.Context, cid string) error {
	cid = fmt.Sprintf("/ipfs/%s", cid)
	p, err := path.NewPath(cid)
	if err != nil {
		return fmt.Errorf("new path by cid error:%w", err)
	}

	err = c.node.Pin().Add(ctx, p)
	if err != nil {
		return fmt.Errorf("add cid error:%w", err)
	}

	return nil
}

// GetInfoByCID 通过cid获取文件信息
func (c *IPFSClient) GetInfoByCID(ctx context.Context, cidStr string) ([]*format.Link, uint64, error) {
	// 判断cid的版本号是否正确
	cc, err := cid.Decode(cidStr)
	if err != nil {
		return nil, 0, fmt.Errorf("decode cid error:%w", err)
	}

	node, err := c.node.Dag().Get(ctx, cc)
	if err != nil {
		return nil, 0, fmt.Errorf("get cid's info error:%w", err)
	}
	size, _ := node.Size()

	return node.Links(), size, nil
}
