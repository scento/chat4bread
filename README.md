# Chat4Bread

Chat4Bread is a hackathon prototype by [Jannik Peters](https://github.com/jannikpeters),
[Mirko Krause](https://github.com/Miroka96) and [Wenzel Pünter](https://github.com/scento) trying
to help local farmers in Cameroon with selling their goods through a SMS-based chatbot marketplace.
This reduces the entry barriers for basic commericial activities. The challenge was sponsored by
[Brot für die Welt](https://www.brot-fuer-die-welt.de/) and [Capgemini](https://www.capgemini.com).

> This is the second iteration of the hackathon project. It contains some major design changes to
> support the testing process and removes dependencies to Twilio which caused financial losses
> at the developers side. Telegram bots will be used to simulate the SMS behavior.

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