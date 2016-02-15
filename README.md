### Requirements
You need to have a registered application for twitter, in order to have a API **key** and a **secret**. Register your application to [apps.twitter.com](https://apps.twitter.com).

Optionnally you can have an notifyMyAndroid key also to have notifications on your phone.

In order to set the parameters for the application, use the following environment variables
* TWITTER_KEY
* TWITTER_SECRET
* TWITTER_USERNAME
* NOTIFY_MY_ANDROID_KEY

### Example
A example of docker command to run: 

```
docker run --rm -it -e TWITTER_KEY=aaaaaaa \
	-e TWITTER_SECRET=bbbbbbbbb \
	-e TWITTER_USERNAME=toto \
	-e NOTIFY_MY_ANDROID_KEY=ccccccc \
	restanrm/listFollowers
```
