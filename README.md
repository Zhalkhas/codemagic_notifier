# CodeMagic Build Notifier

This is Telegram bot that sends build artifacts at post-publish stage.

# Setup

Create new project at Google Cloud Platform and enable Google Cloud Build API.
Detailed instructions can be found [here](https://cloud.google.com/functions/docs/create-deploy-gcloud#functions-prepare-environment-go)

To deploy a function, execute the following command:

```bash
  gcloud functions deploy codemagic_notifier --env-vars-file .env.yaml --runtime go119 --trigger-http --allow-unauthenticated
``` 

Example of .env.yaml can be found at the root of the project.

