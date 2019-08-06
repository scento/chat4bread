# Chat4Bread

Chat4Bread is a hackathon prototype by [Jannik Peters](https://github.com/jannikpeters),
[Mirko Krause](https://github.com/Miroka96) and [Wenzel Pünter](https://github.com/scento) trying
to help local farmers in Cameroon with selling their goods through a SMS-based chatbot marketplace.
This reduces the entry barriers for basic commericial activities. The challenge was sponsored by
[Brot für die Welt](https://www.brot-fuer-die-welt.de/) and [Capgemini](https://www.capgemini.com).

> This is the second iteration of the hackathon project. It contains some major design changes to
> support the testing process and removes dependencies to Twilio which caused financial losses
> at the developers side. Telegram bots will be used to simulate the SMS behavior.

## Features

- Onboarding
- Find local farmers nearby
- Sell a product
- Buy a product
- Quote average market prices for a product

## Known Bugs

- **Intent Misclassification**: As we are using very few sentence samples and do not separate the onboarding and usage steps, it frequently happens that the bot misinterprets requests. Well known cases include the question "Are you a farmer or consumer?".
- **Missing Distance Metric**: We currently do not enforce the geoposition of offers. In consequence, it is possible to find a trade at the other end of the planet.
- **Self-Trading**: It is possible to trade with the own account. This was deliberately left over for debugging purposes but might be strange.
- **Database Maintancence**: Offers are not removed from the database even when everything is sold.
- **No trade aggregation**: We are only matching a single trade without the possibility to combine multiple vendors.
- **No confirmations**: All actions are performed directly without intermediate acceptance questions, which might cause bugs.
- **No data maintenance**: It is only possible to remove or change personal data by contacting the database administrator.
- **SMS Security**: When using SMS, it is possible to fake these within the network. Relevant actions should require a password.

## Deployment

Obtain a [Telegram bot token](https://www.siteguarding.com/en/how-to-get-telegram-bot-api-token)
and a SAP Conversational AI token for the [NLP project](https://cai.tools.sap/scento/chat4bread).
Think of a MongoDB username and password (e.g. with `pwgen`) for your setup. Then execute the
following steps in the project root, assuming [Docker](https://www.docker.com/) is installed and
running:

```bash
export MONGO_USERNAME={your MongoDB username}
export MONGO_PASSWORD={your MongoDB password}
export TELEGRAM_TOKEN={your Telegram bot token}
export CAI_TOKEN={your SAP CAI token}

docker-compose up
```

If you did no changes to the NLP project, you can use the token `d363362493ea638ec0a529773316feec`
for CAI.

To keep the bot running in the background use `docker-compose up -d` as the final statement.

When the bot is started for the first time a new docker volume will be created to store the data of the Mongo DB. Therefore, the containers can be recreated as necessary. To empty the database shut down the containers and remove the volume. You can get its name using `docker volume ls` and search for `..._mongo_data`. After getting the name, the volume can be removed using `docker volume rm [volume name]`.

### Using Telegram Web Hooks
You can also use telegram web hooks. Using those, Telegram will perform a REST request to your server as soon as a new message arrives. This method replaces the default long polling approach, where the backend (a.k.a. the bot) performs many iterative requests no matter if there are new messages or not. Therefore, web hooks are less resource intensive, but you need a publicly reachable, SSL-secured (https) website for this. There should be a reverse proxy serving HTTPS and (e.g. apache or nginx) forwarding requests to your bot at http://localhost:8081 (you can change the port in the docker-compose.yml). You should use a firewall to prevent direct external requests to the bot or modify the port forwarding in the docker-compose.yml to only listen on localhost.

To enable web hooks, call `export TELEGRAM_WEBHOOK_URL=chat4bread.example.com/` before calling `docker-compose up`:

```bash
export TELEGRAM_WEBHOOK_URL=chat4bread.example.com/
```

A sample nginx reverse proxy configuration may look like this:

```
# cat /etc/nginx/sites-enabled/chat4bread.example.com
server {
  server_name chat4bread.example.com;

  location / {
    proxy_pass	http://localhost:8081;
    proxy_set_header Host $http_host;
  }

  listen [::]:443;
  listen 443;
  ssl_certificate ...;
  ssl_certificate_key ...;
}

```

## Debugging
### Getting direct database access
```
docker exec -it chat4bread-database /bin/bash
$ mongo
> use admin
> db.auth("{your MongoDB username}", "{your MongoDB password}")
> use chat4bread
> show collections
> db.users.find()
...
```
