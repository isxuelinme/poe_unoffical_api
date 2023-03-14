package cli

import (
	"bufio"
	"fmt"
	"github.com/inancgumus/screen"
	"github.com/isxuelinme/poe_unoffical_api/core"
	"os"
)

// CLI it is simplest cli implement for poe_unoffical_api
// just for noobs asking me ...:)
func CLI() {
	core.SetLogMode(core.LOG_ERROR)
	mutilUser := core.NewMutLtiUserGpt(core.GptTypePoeUnofficial)
	reader := bufio.NewReader(os.Stdin)
	screen.Clear()
	screen.MoveTopLeft()

	AllQuestionAndAnswer := ""

	messageEndChan := make(chan bool)

	fmt.Print(" Question: ")
loop:
	text, _ := reader.ReadString('\n')
	AllQuestionAndAnswer += " Question: " + text
	lastMessageId := ""
	tempAnswer := " Answer: "
	mutilUser.Talk(&core.AskRequest{
		UserId:           1,
		Question:         text,
		CallbackFuncName: "",
		AskResponseCallBack: func(askRequest *core.AskRequest, message *core.CallbackMessageResponse) {
			if lastMessageId == message.Data.Id {
				return
			}
			if message.Data.IsEnd {
				screen.Clear()
				screen.MoveTopLeft()
				fmt.Printf("\r%s", AllQuestionAndAnswer+tempAnswer+message.Data.Text+"\n Question:")
				AllQuestionAndAnswer += tempAnswer + message.Data.Text + "\n"
				//fmt.Printf("\r answer: %s\n question:", message.Data.Text)
				lastMessageId = message.Data.Id
				messageEndChan <- true
				return
			}
			screen.Clear()
			screen.MoveTopLeft()
			fmt.Printf("\r%s", AllQuestionAndAnswer+tempAnswer+message.Data.Text)
		},
	})
	<-messageEndChan
	goto loop
}
