# Tell | Free push notifications from the command line, delivered via Telegram

Work in progress, incomplete. Most things described here don't work yet.

## Roadmap

- [x] loading configuration, getting the bot token and conversation ID
- [x] sending basic messages
- [x] authorizing users
- [x] handling standard input
- [ ] sending small files
- [ ] handling voice messages, photos and songs
- [ ] handling the uploading of large files
- [ ] interactive setup
- [ ] Receiving files and messages on demand
- [ ] Background execution
- [ ] executing predefined commands
- [ ] Saving files
- [ ] sending notifications with buttons
- [ ] securely receiving reactions
- [ ] handling arbitrary commands
- [ ] automatic configuration transfer
- [ ] Multiple conversations, conversation aliases.
- [ ] Group support.

## What is this?

Tell lets you send push notifications to all your devices with one simple command. It also supports receiving messages, files and (either predefined or arbitrary) commands back. This can be used to:

- easily send push notifications from your shell scripts, for free
- notify you and your team when your cron jobs succeed or fail
- act as a magic wormhole replacement, except faster and with no pesky codes
- let users react to notifications by securely executing predefined commands with one click
- inform you about problems with your server and applications, suspicious activity and failed login attempts
- send a message to your phone when a long running task has finished
- send custom notifications to a small group of users, no matter the platform, for free
- let you securely start common tasks on your quickly, no matter where you are
- let you interact with your computer on the go by executing arbitrary commands (disabled by default)

## Setup

### Installation

Tell is distributed as a single binary. Download it, put it on your path, and that's it.

## Simple setup

After Tell is installed and on your path, just run `tell` and follow the displayed instructions.

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

Send a simple message, Tell works just like `echo`:

```bash
tell Hello, World!
```

Send the contents of STDIN as a text message:
```bash
tell < path/to/your/file.txt
ps -aux | head -n 10 | tell
````

### Sending files:

Send a normal file (will be sent as a document):

```bash
tell myfile.exe
```

If your file has the right extension, it will be sent as a photo, voice message etc.

```bash
tell -f hello.jpg 'Photo caption'
````

Telegram has pretty strict limitations on what is allowed in those messages. If sending fails, your file will be sent as a normal document.

You can also send a folder. It will be zipped up and sent as a document:

```bash
tell -f my_work
```

You can manually specify a file type with the `-t` flag:

```bash
tell -t=document file.mp3
```

Telegram bots aren't allowed to send files larger than 50MB. Instead, those files will be uploaded to [transfer.sh](https://transfer.sh). The URL to the uploaded file will be sent as a 
text message. 

You can encrypt an uploaded file or folder with the `e` flag:
```
tell -ef credit_card_details.txt
```

Encrypted files can only be sent as "documents" or uploaded to transfer.sh. They can only be decrypted on another computer running Tell. The two computers must have the same secret key configured in `~/.tell.json`. This happens automatically when the config transfer mechanism is used.

### Receiving messages and files

You can use Tell to transfer data between two computers, or from another device with Telegram access. After sending a message to the bot, either via Tell itself or Telegram, you can do:

```
tell -r
```

and the message will be displayed. If you send a file, the file will be downloaded and saved in your current working directory. Encrypted files are decrypted automatically, as long as the key matches. Folders are automatically unzipped. Voice messages, photos etc. are saved under an automatically generated name. If the last message contains nothing but a single URL, the file under that URL is downloaded. If the message contains one or more URLs, possibly interspersed with other text, you can pass `-u=1,2,4` to download the first, second and fourth URL. Pass `-u=all` to download everything. This flag doesn't have any effect for non-text messages and messages without URLs, so it can be safely passed each time you run `tell -r`.

### Executing commands from the user:

To execute commands from Telegram, Tell needs to run in the background and wait for new messages. Use the `tell -d` command to start it up. 


#### Handling predefined commands:

By default, Tell will only let you execute scripts from the `tellscripts` folder located in your home (or user) directory. This limitation prevents potential hackers from wreaking havoc on your server by hacking your Telegram account. This folder is usually located at `/home/<your_username>/tellscripts` on Linux, `C:\Users\<your_username>\tellscripts` on Windows, and `/Users/<your_username>/tellscripts` on Mac. If it doesn't exist, you can just create it yourself. For example, if this directory contains a script called `test.py` (the extension doesn't matter), you can execute it with `/test`. If you're using Linux or Mac OS, remember to put a comment like `#!/bin/env python3`  (or equivalent for other languages) in the first line of your script, see [This article on the "Shebang"](https://bash.cyberciti.biz/guide/Shebang) for more info. If you want to allow an existing command like `ls` or `cat`, you can create a symbolic link with:

```bash
ln -s path_to_command ~/tellscripts/command_name_to_use
```

To figure out where a command such as `ls` is located, do:

```
which ls
```

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

You can add extra buttons to the notifications you send. Clicking those buttons will cause commands to be executed. You can use this feature to let users quickly restart failed builds,  ask for more information etc. This is implemented in a secure way, users can't abuse this feature to run arbitrary commands on your server.

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

If you want to specify who to send a notification to, use the `-t` flag, followed by one or more recipients separated by commas. Your command should look something like the following:

```bash
tell -t admin1,admin2,admin3 We have an issue, come fix it.
```

## File uploads

All files larger than 50mb are uploaded to (transfer.sh)[transfer.sh] and sent as links. Mp3 and m4a files are send as audio (music) files, which are different from voice messages. Ogg files are send as voice messages; they must be encoded with the opus codec, **NOT** the Vorbis codec. Jpg and png files smaller than 10MB are send as photos. Their width and height must not exceed 10000 in total, and the ratio of width and height must not be larger than 20. Photos larger than 10MB are uploaded to transfer.sh, while photos with a wrong width, height or ratio can't be uploaded at all. Gif files are sent as animations. All other files are uploaded as documents. Files are never uploaded as video notes or stickers by default.

You can manually change the file type with the `--file-type` flag. The allowed values are `animation`, `audio`, `document`, `photo`, `sticker`, `video`, `video_note`, `voice` and `upload`. If you use the `--file-type` flag, you must make sure that your file follows all of Telegram's restrictions for that type. You can check what those restrictions are in the [Telegram bot API documentation](https://core.telegram.org/bots/api). The `upload` file type is specific to Tell, files of this type are uploaded to transfer.sh and sent as links.
