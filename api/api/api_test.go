package api_test

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/koeng101/dnadesign/api/api"
)

var app api.App

func TestMain(m *testing.M) {
	app = api.InitializeApp()
	code := m.Run()
	os.Exit(code)
}

func TestIoFastaParse(t *testing.T) {
	baseFasta := `>gi|5524211|gb|AAD44166.1| cytochrome b [Elephas maximus maximus]
LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLV
EWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLG
LLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVIL
GLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGX
IENY
`
	req := httptest.NewRequest("POST", "/api/io/fasta/parse", strings.NewReader(baseFasta))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	r := `[{"identifier":"gi|5524211|gb|AAD44166.1| cytochrome b [Elephas maximus maximus]","sequence":"LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLVEWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLGLLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVILGLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGXIENY"}]`
	if strings.TrimSpace(resp.Body.String()) != r {
		t.Errorf("Unexpected response. Expected: " + r + "\nGot: " + resp.Body.String())
	}
}
