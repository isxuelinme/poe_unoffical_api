# Using POE Unofficial API

## Import

```dotenv
go get github.com/isxuelinme/poe_unofficial_api/core@last
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

## More details in core and example

```golang
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
```

## It's easy to use, but I can't open the source SSE (http event stream) now. Maybe later. However, you can use AskResponseCallBack to implement it by yourself.

## It has implemented multi-user, but it is not friendly to business and especially noobs, just for dev/test. So you have to read the code by yourself.