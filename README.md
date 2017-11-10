# Escrow bot

#### Available commands
```js
"/deposit"       [POST]  //step(int), from(string), to(string), amount(int)
"/pay"           [POST]  //step(int)
"/refund"        [POST]  //step(int)
"/confirmations" [POST]  //tx(string)
```
Don't forget to include http basic-auth header
```js
Authorization Basic cbGad1XVHxFL
```