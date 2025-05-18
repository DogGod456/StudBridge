package database

const Schema = `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Типы участников
CREATE TYPE message_status AS ENUM ('sent', 'delivered', 'read', 'failed');

-- Таблица типов участников
CREATE TABLE IF NOT EXISTS participant_types (
    id_participant_type UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Таблица участников
CREATE TABLE IF NOT EXISTS participants (
    id_participant UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_participant_type UUID NOT NULL,
    ref_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Таблица чатов
CREATE TABLE IF NOT EXISTS chats (
    id_chat UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Таблица участников чата
CREATE TABLE IF NOT EXISTS chat_senders (
    id_chat_sender UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL,
    id_participant UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    UNIQUE(id_chat, id_participant)
);

-- Таблица сообщений
CREATE TABLE IF NOT EXISTS messages (
    id_message UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL,
    id_sender UUID NOT NULL,
    message_text TEXT NOT NULL,
    status message_status NOT NULL DEFAULT 'sent',
    id_reply_message UUID,
    draft BOOLEAN NOT NULL DEFAULT FALSE,
    sending_time TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Связи между таблицами
ALTER TABLE participants
ADD CONSTRAINT fk_participant_type
FOREIGN KEY (id_participant_type)
REFERENCES participant_types(id_participant_type);

ALTER TABLE chat_senders
ADD CONSTRAINT fk_chat_sender_id_chat
FOREIGN KEY (id_chat)
REFERENCES chats(id_chat)
ON DELETE CASCADE;

ALTER TABLE chat_senders
ADD CONSTRAINT fk_chat_sender_id_participant
FOREIGN KEY (id_participant)
REFERENCES participants(id_participant);

ALTER TABLE messages
ADD CONSTRAINT fk_messages_id_chat
FOREIGN KEY (id_chat)
REFERENCES chats(id_chat) 
ON DELETE CASCADE;

ALTER TABLE messages
ADD CONSTRAINT fk_messages_id_sender
FOREIGN KEY (id_sender)
REFERENCES participants(id_participant);

ALTER TABLE messages
ADD CONSTRAINT fk_messages_id_reply_message
FOREIGN KEY (id_reply_message)
REFERENCES messages(id_message);

`
