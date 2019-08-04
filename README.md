# Chat4Bread

Chat4Bread is a hackathon prototype by Jannik Peters, Mirko Krause and Wenzel Pünter trying to help local farmers in Cameroon with
selling their goods through a SMS-based chatbot marketplace. This reduces the entry barriers for basic commericial activities.

> This is the second iteration of the hackathon project. It does some major design changes to
> support the testing process and removes dependencies to Twilio which caused financial losses
> at the developers side. Instead, Telegram bots will be used to simulate the SMS behavior.

## Deployment

Obtain a [Telegram bot token](https://www.siteguarding.com/en/how-to-get-telegram-bot-api-token) and a SAP Conversational AI token for the [NLP project](https://cai.tools.sap/scento/chat4bread). Think of a MongoDB username and password (e.g. with `pwgen`) for your setup. Then execute the following steps in the project root, assuming [Docker](https://www.docker.com/) is installed and running:

```bash
export MONGO_USERNAME={your MongoDB username}
export MONGO_PASSWORD={your MongoDB password}
export TELEGRAM_TOKEN={your Telegram bot token}
export CAI_TOKEN={your SAP CAI token}

docker-compose up
```
