package main

import (
	"fmt"
	"github.com/isxuelinme/poe_unoffical_api/core"
)

func main() {
	mutilUser := core.NewMutLtiUserGpt(core.GptTypePoeUnofficial)
	ask := &core.AskRequest{
		UserId:           0,
		Question:         "hi~ bro",
		CallbackFuncName: "",
		AskResponseCallBack: func(askRequest *core.AskRequest, response []byte) {
			fmt.Println(string(response))
		},
	}
	mutilUser.Talk(ask)

	select {}

}
