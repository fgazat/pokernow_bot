# Pokernow.club telegram bot

Bot: [@pokernowclub_bot](https://t.me/pokernowclub_bot)

Supported commands:

* `/calc URL` — posts comment with transactions.
* `/map IN_GAME_NICKNAME TG_LOGIN` — appends nickname to existing UserInfo entry.
* `/new IN_GAME_NICKNAME TG_LOGIN PAYMENT_INFO` — creates new UserInfo entry.

TODO:

* `/my_payments` — return list of transactions — we have to store info about every game and user and somehow check if payments done.
    File based database? psql?
    Should work in private messages.
* support multiroom processing — for every chatID create its []UserInfo and process it. 
* summary html page for every room (chatID) with list of games and links to full result page.
