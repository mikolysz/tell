# Tell | Free push notifications from the command line, delivered via Telegram

Work in progress, incomplete. Most things described here don't work yet.

## Roadmap

- [ ] loading configuration, getting the bot token and conversation ID
- [ ] sending basic messages
- [ ] Manually setting the bot token
- [ ] authorizing users
- [ ] handling standard input
- [ ] sending small files
- [ ] handling voice messages, photos and songs
- [ ] handling the uploading of large files
- [ ] background execution, registering the webhook
- [ ] figuring out if we're exposed or not
- [ ] executing predefined commands
- [ ] sending notifications with buttons
- [ ] securely receiving reactions
- [ ] handling arbitrary commands

## What is this?

Tell lets you send push notifications to all your devices with one simple command. It also supports receiving (either predefined or arbitrary) commands back. This can be used to:

- easily send notifications from your shell scripts
- notify you and your team when your cron jobs succeed or fail
- let users react to notifications by executing predefined commands with one click
- inform you about problems with your server and applications, suspicious activity and failed login attempts
- send a message to your phone when a long running task has finished
- send custom notifications to a small group of users, no matter the platform, for free
- let you start common tasks quickly, no matter where you are
- let you interact with your computer on the go by executing arbitrary commands (disabled by default)

## Setup

### Installation

Tell is distributed as a single binary. Download it, put it on your path, and that's it.

## #Creating a Telegram bot

To use Tell, you need to create a Telegram bot and Obtain a bot token. Here's how to do it:

1. Message [@BotFather](https://t.me/botfather). The command you should send is '/newbot'.
2. You will be asked to give your bot a name and an username.
3. When you do so, you will receive a token, which you will need in the next step.

### Setting up tell.

You can set your bot token by entering the following command:

```bash
tell -t <your_bot_token>
````

You also need to tell your bot who to send notifications to. To do so, use the `tell -a` command. An authorization code will be displayed. Send this code to the bot as a Telegram message, and you should be ready to send your notifications. Tell will show you a conversation ID, you don't need it unless you plan to support multiple users.

## Usage

Send a simple message:

```bash
tell Hello, World!
```

Send the contents of a text file:
```bash
tell < path/to/your/file.txt
````

Send one or more binary files, for example pictures and audio files:
```bash
tell -f photos/picture1.jpg -f music/song.mp3
```

Audio files are sent as voice messages. Image files are sent as Telegram photos. Other files smaller than 50MB are sent via Telegram. Bigger files are uploaded to [transfer.sh](transfer.sh), and a link to the file is sent. If you want to send audio files as music instead, use the `-m` option instead of `-f`. This lets you play these files in Telegram's music player, which supports features such as next/previous song.

### Executing commands from the user:

To execute commands from Telegram, Tell needs to run in the background and wait for new messages. Use the `tell -b` command to start it up. 

**Note**: For this to work, you need to expose port 2137 to the world. This is usually not an issue on servers, but might be a problem on home computers behind a router.

### Handling predefined commands:

By default, Tell will only let you execute scripts from the `tellscripts` folder located in your home (or user) directory. This limitation prevents potential hackers from wreaking havoc on your server by hacking your Telegram account. This folder is usually located at `/home/<your_username>/tellscripts` on Linux, `C:\Users\<your_username>\tellscripts` on Windows, and `/Users/<your_username>/tellscripts` on Mac. If it doesn't exist, you can just create it yourself. If this directory contains a script called `test.py` (the extension doesn't matter), you can execute it with `/test`. If you're using Linux or Mac OS, remember to put a comment like `#!/bin/env python3`  (or equivalent for other languages) in the first line of your script, see [This article on the "Shebang"](https://bash.cyberciti.biz/guide/Shebang) for more info.

If you pass arguments to your Telegram command, those arguments will be passed straight to your script. So, a command like:
```
/myscript --option value arg1 arg2 arg3
```

Is equivalent to the following Bash command (assuming there's a myscript.sh file in your telscripts directory):

```bash
~/telscripts/myscript.sh --option value arg1 arg2 arg3
````

**Note**: Don't put two files with the same name but different extensions in the "telscripts" folder, or strange things will happen.

### Notification reactions

You can add buttons to your notifications. Clicking those buttons will cause commands to be executed. You can use this feature to let users quickly restart failed builds,  ask for more information etc.

To send a notification that includes such buttons, use a command like the following:

```bash
tell -o 'Repeat build: gcc -o hello hello.c' -o 'More info: tell -f error.log' Build failed.
````

This will send a message saying "build failed" with two buttons, "Repeat build" and "More info".

The argument after '-o' should always be surrounded by single quotes (or double quotes if you're on Windows). The button name and the command to execute are separated by a colon. Whitespace between the colon and the command is removed.

### Executing arbitrary commands

If you want to let users execute arbitrary commands, start Tell in the background like this:

```bash
tell -b --danger-allow-arbitrary-commands-i-know-what-i-am-doing
````

Then, if an authorized user sends a command to the bot that isn't prefixed by a slash, the command gets executed and the output is sent back as a message.

## Handling multiple users

You can use the `tell -a` command multiple times to add more than one user. All notifications are sent to all authorized users by default.

Each user you add is assigned an internal ID, which is supposed to be short and memorable. If you want to use a custom ID, authorize the user like this

```bash
tell -a --id my_custom_id
````

If you want to specify who to send a notification to, use the `-r` flag, followed by one or more recipients separated by commas. Your command should look something like the following:

```bash
tell -r admin1,admin2,admin3 We have an issue, come fix it.
```

