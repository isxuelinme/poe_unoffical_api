# using POE unofficial api
### import 
```dotenv
go get github.com/isxuelinme/poe_unoffical_api/core
```
### Run the following code on your chrome console 
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
### Copy the value of POV_CHANNEL and POV_CHAT_ID, outputs like following 
```dotenv
POV_CHANNEL =  poe-chan51-8888-hhmp2zuksgonnzdwnitj
POV_CHAT_ID =  550223
```

### Change .env.example name to .env and change the value of cookie of yourself
```dotenv
POE_COOKIE = <your cookie>
POV_CHANNEL = <your channel>
POV_CHAT_ID = <your chat_id>
```
### More detail in example
### Not friendly to business and especially noobs,
### just for dev/test
