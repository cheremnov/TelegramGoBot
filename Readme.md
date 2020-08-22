A Telegram bot. Can repeat your words and notify you. Also checks your authorization.

Current features:
1) Reminding about future events. 

Future features:
1) Reminding about the next lecture. A classroom, a lecturer's name and other useful things.
2) Accessing notes repository.
3) Unit and integration bot testing


How to create a bot copy. Assuming mysql installed:
1. Fork the repository
2. Create config.json. Add to the "Token" field your bot token.
3. Create keywords.json. Add to the "Keyword" field your keyword. 
4. Set up a database
5. Launch the bot. 
6. (optional) Delete keywords.json. The bot has generated a salted keyword, it doesn't need a plain-text keyword.

How to set up a database:
1. Create a "TelegramBot" user in your database.
2. (optional) Change a database name in the sql scripts.
3. Create a database with a default "WaspDatabase" or your chosen name.
4. Execute a script creating the tables.
5. Set a databaseConfig.json. Set a TelegramBot user password. Set your server and database name.
Add the Internet protocol used to connect to the database server. (ProtocolType)

The bot configuration constants:
kNotificationsRefreshTime — how often look for expired notifications
kUTCTimeZone — a timezone
