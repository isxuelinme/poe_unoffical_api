package core

var payLoadForTalk = `
	{
  "operationName": "AddHumanMessageMutation",
  "query": "mutation AddHumanMessageMutation($chatId: BigInt!, $bot: String!, $query: String!, $source: MessageSource, $withChatBreak: Boolean! = false) {\n  messageEdgeCreate(\n    chatId: $chatId\n    bot: $bot\n    query: $query\n    source: $source\n    withChatBreak: $withChatBreak\n  ) {\n    __typename\n    message {\n      __typename\n      node {\n        __typename\n        ...MessageFragment\n        chat {\n          __typename\n          id\n          shouldShowDisclaimer\n        }\n      }\n    }\n    chatBreak {\n      __typename\n      node {\n        __typename\n        ...MessageFragment\n      }\n    }\n  }\n}\nfragment MessageFragment on Message {\n  id\n  __typename\n  messageId\n  text\n  linkifiedText\n  authorNickname\n  state\n  vote\n  voteReason\n  creationTime\n  suggestedReplies\n}",
  "variables": {
    "bot": "capybara",
    "chatId": 550922,
    "query": "现在还记得吗？\n",
    "source": null,
    "withChatBreak": false
  }
}
`

var payLoadForGetHistory = `
	{
  "operationName": "ChatPaginationQuery",
  "query": "query ChatPaginationQuery($bot: String!, $before: String, $last: Int! = 10) {\n  chatOfBot(bot: $bot) {\n    id\n    __typename\n    messagesConnection(before: $before, last: $last) {\n      __typename\n      pageInfo {\n        __typename\n        hasPreviousPage\n      }\n      edges {\n        __typename\n        node {\n          __typename\n          ...MessageFragment\n        }\n      }\n    }\n  }\n}\nfragment MessageFragment on Message {\n  id\n  __typename\n  messageId\n  text\n  linkifiedText\n  authorNickname\n  state\n  vote\n  voteReason\n  creationTime\n  suggestedReplies\n}",
  "variables": {
    "before": null,
    "bot": "capybara",
    "last": 20
  }
}
`
