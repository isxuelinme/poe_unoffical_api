package main

import (
	"fmt"
	"github.com/isxuelinme/poe_unoffical_api/core"
)

func main() {
	core.SetLogMode(core.LOG_ERROR)
	MutLtiUser := core.NewMutLtiUserGpt(core.GptTypePoeUnofficial)
	ask := &core.AskRequest{
		UserId:           1,
		Question:         "hi~ bro",
		CallbackFuncName: "",
		AskResponseCallBack: func(askRequest *core.AskRequest, response *core.CallbackMessageResponse) {
			fmt.Println(response.text)
		},
	}
	//ask question
	MutLtiUser.Talk(ask)

	select {}

}
