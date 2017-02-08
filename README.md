# EmailService

This is a MicroService written in Go to send Email via SMTP. It is using Amazon SES at the moment but can use any SMTP email server. The installation process is detailed in the dev folder within the Docker file. This is the easiest way to deploy.

To Send an Email:

Post: URL/email/send
JSON:
{
"from":"from email",
"to":"to email",
"subject":"subject line",
"body":"email contents"
}
