/* Only the authorized users */
/* A user table is redundant, Telegram provides this information*/
CREATE TABLE WaspDatabase.Notifications
(
notification_id INTEGER PRIMARY KEY AUTO_INCREMENT,
user_id INTEGER,
message VARCHAR(100) CHARACTER SET utf8,
notification_time DATETIME
);
/* A table connecting the user to his or her private chat */
/* Assume only one private chat */
CREATE TABLE WaspDatabase.UserPrivateChats
(
user_id INTEGER PRIMARY KEY,
chat_id INTEGER
);
