## USING  SSE
https://user-images.githubusercontent.com/13929427/225206524-36ec3a9f-bf6a-4252-8cd0-5deeef93da7c.mp4

### config env & RUN & and open [http://localhost:8090](localhost:8090/events)
```dotenv
RUN_MODE = SSE
BACKEND_PORT = 8090
```
### code just is explanation
#### you can use following code to implement yourself web ui 
```javascript
    //pack message to json for AskRequest function
    function newMessage(text){
        return {
            type: 'conversation',
            conversation_id: '',
            parent_message_id: "",
            text: text
        }
    }

    //ask or send message and waiting following SSE response
    function A(data){
        fetch('/conversation', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(newMessage(data))
        }).then(function(response) {
            return response.json();
        }).then(function(data) {
            console.log(data);
        });
    }

    class SSE {
        constructor(url) {
            this.url = url;
            this.eventSource = new EventSource(url);
            this.eventSource.addEventListener('message', this.onMessage.bind(this));
            this.eventSource.onerror = this.onError.bind(this);
            this.eventSource.onopen = this.onOpen.bind(this);

            this.eventSource.addEventListener('done', this.onDone.bind(this));
        }
        //when message come, this function will be called
        //the event.data is plain text not json format
        onMessage(event) {
            console.clear()
            console.log("casue I use console.clear(),the console refresh, dont worry it\nmessage",event.data);
        }
        onError(event) {
            console.log("error",event);
        }
        onOpen(event) {
            console.log("open",event);
        }
        //when message done, this function will be called but not 1 time maybe more times it contain suggestion 
        //u can process end of UI logic here
        //the event.data is JSON format
        onDone(event) {
            console.log("done",event.data);
        }

    }
    e = new SSE('/events');


```
