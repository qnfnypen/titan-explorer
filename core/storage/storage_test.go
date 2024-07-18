package storage

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gnasnik/titan-explorer/pkg/random"
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
