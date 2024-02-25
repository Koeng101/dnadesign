package api

import (
	"testing"

	"github.com/koeng101/dnadesign/api/gen"
)

func TestApp_LuaIoFastaParse(t *testing.T) {
	luaScript := `
parsed_fasta = fasta_parse(attachments["input.fasta"])

output = parsed_fasta[1].identifier
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

	_, output, err := app.ExecuteLua(luaScript, []gen.Attachment{fastaAttachment})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := "AAD44166.1"
	if output != expectedOutput {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
	}
}
