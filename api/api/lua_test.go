package api_test

import (
	"testing"

	"github.com/koeng101/dnadesign/api/api"
	"github.com/koeng101/dnadesign/api/gen"
)

func TestApp_Examples(t *testing.T) {
	output, err := app.ExecuteLua(api.Examples, []gen.Attachment{})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := `test
GATC
`
	if output != expectedOutput {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
	}
}

func TestApp_LuaIoFastaParse(t *testing.T) {
	luaScript := `
parsedFasta = fastaParse(attachments["input.fasta"])

print(parsedFasta[1].identifier)
`
	inputFasta := `>AAD44166.1
LCLYTHIGRNIYYGSYLYSETWNTGIMLLLITMATAFMGYVLPWGQMSFWGATVITNLFSAIPYIGTNLV
EWIWGGFSVDKATLNRFFAFHFILPFTMVALAGVHLTFLHETGSNNPLGLTSDSDKIPFHPYYTIKDFLG
LLILILLLLLLALLSPDMLGDPDNHMPADPLNTPLHIKPEWYFLFAYAILRSVPNKLGGVLALFLSIVIL
GLMPFLHTSKHRSMMLRPLSQALFWTLTMDLLTLTWIGSQPVEYPYTIIGQMASILYFSIILAFLPIAGX
IENY
`

	fastaAttachment := gen.Attachment{
		Name:    "input.fasta",
		Content: inputFasta,
	}

	output, err := app.ExecuteLua(luaScript, []gen.Attachment{fastaAttachment})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := "AAD44166.1\n"
	if output != expectedOutput {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
	}
}
