# Using POE Unofficial API

## demo SSE [READ SSE README.md](https://github.com/isxuelinme/poe_unoffical_api/blob/main/client/sse/README.md)

https://user-images.githubusercontent.com/13929427/225206524-36ec3a9f-bf6a-4252-8cd0-5deeef93da7c.mp4


## demo CLI

https://user-images.githubusercontent.com/13929427/224995915-ba5f873f-28ab-4dec-8790-a760c0dcc547.mp4

## Import

```dotenv
import "github.com/isxuelinme/poe_unofficial_api/core"
go mod tidy 
```

## Run the following code on your Chrome console

```javascript
function getChatId() {
    let channel = localStorage.getItem("poe-tchannel-channel")
    let paramsForGetChatId = window.__NEXT_DATA__.buildId
    let fetchUrl = "https://poe.com/_next/data/" + paramsForGetChatId + "/sage.json?handle=sage"
    fetch(fetchUrl)
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not okay');
            }
            return response.text();
        })
        .then(data => {
            jsonData = JSON.parse(data)
            console.log("POV_CHANNEL = ", channel)
            console.log("POV_CHAT_ID = ", jsonData.pageProps.payload.chatOfBotDisplayName.chatId)
        })
        .catch(error => {
            console.error('Error fetching data:', error);
        });
}(getChatId())
```

## Copy the value of POV_CHANNEL and POV_CHAT_ID after running the above code. The output will look like this:

```dotenv
POV_CHANNEL =  poe-chan51-8888-hhmp2zuksgonnzdwnitj
POV_CHAT_ID =  550223
```

## Change .env.example name to .env and change the value of your cookie
```dotenv
POE_COOKIE = <your cookie>
POV_CHANNEL = <your channel>
POV_CHAT_ID = <your chat_id>
```
## if you wanna use SSE (default is CLI), ADD the following configuration to .env
```dotenv
RUN_MODE = SSE
BACKEND_PORT = <backend port if not set default is 6000>
```

## More details in core and example

```golang
package main
import "github.com/isxuelinme/poe_unofficial_api/core"

func main() {
    core.SetLogMode(core.LOG_ERROR)
    MutLtiUser := core.NewMutLtiUserGpt(core.GptTypePoeUnofficial)
    ask := &core.AskRequest{
        UserId:           1, //your local user id
        Question:         "hi~ bro",
        CallbackFuncName: "", //useless. Like JSONP or event name
       // when message Coming from GPT, it will call this function
        AskResponseCallBack: func(askRequest *core.AskRequest, response *core.CallbackMessageResponse) {
			fmt.Printf("\r answer: %s", message.Data.Text)
        },
    }
    //ask question
    MutLtiUser.Talk(ask)
    
    select {}
}
```

## It's easy to use,However, you can use AskResponseCallBack to implement websocket or more protocol by yourself. learn more [READ SSE IMPLEMENT](http://github.com/isxuelinme/poe_unofficial_api/client/sse/SSE.go)
```golang
 ask := &core.AskRequest{
        UserId:           1,
        Question:         "hi~ bro",
        CallbackFuncName: "",
        AskResponseCallBack: func (askRequest *core.AskRequest, response *core.CallbackMessageResponse) {
			<your business logiical code>
        },
 }
```
## It has implemented multi-user, but it is not friendly to business and especially noobs, just for dev/test. So you have to read the code by yourself.
