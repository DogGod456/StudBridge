package database

const Schema = `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Типы участников
CREATE TYPE message_status AS ENUM ('sent', 'delivered', 'read');

-- Таблица типов участников
CREATE TABLE IF NOT EXISTS participant_types (
    id_participant_type UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type_name VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Таблица участников
CREATE TABLE IF NOT EXISTS participants (
    id_participant UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_participant_type UUID NOT NULL REFERENCES participant_types(id_participant_type),
    ref_id TEXT NOT NULL, -- ID из микросервиса Networking
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
    id_chat UUID NOT NULL REFERENCES chats(id_chat) ON DELETE CASCADE,
    id_sender UUID NOT NULL,
    id_participant UUID NOT NULL REFERENCES participants(id_participant),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    UNIQUE(id_chat, id_participant)
);

-- Таблица сообщений
CREATE TABLE IF NOT EXISTS messages (
    id_message UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL REFERENCES chats(id_chat) ON DELETE CASCADE,
    id_sender UUID NOT NULL,
    message_text TEXT NOT NULL,
    status message_status NOT NULL DEFAULT 'sent',
    id_parent_message UUID,
    draft BOOLEAN NOT NULL DEFAULT FALSE,
    sending_time TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Таблица папок чатов
CREATE TABLE IF NOT EXISTS chats_folders (
    id_chats_folder UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_participant UUID NOT NULL REFERENCES participants(id_participant),
    folder_name VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- Таблица чатов в папках
CREATE TABLE IF NOT EXISTS chats_in_folder (
    id_chats_in_folder UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL REFERENCES chats(id_chat) ON DELETE CASCADE,
    id_chats_folder UUID NOT NULL REFERENCES chats_folders(id_chats_folder) ON DELETE CASCADE,
    UNIQUE(id_chat, id_chats_folder)
);

-- Таблица архивных чатов
CREATE TABLE IF NOT EXISTS archive_chat (
    id_archive_chat UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL REFERENCES chats(id_chat) ON DELETE CASCADE,
    id_participant UUID NOT NULL REFERENCES participants(id_participant),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(id_chat, id_participant)
);

-- Таблица закреплённых чатов
CREATE TABLE IF NOT EXISTS pinned_chat (
    id_pinned_chat UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL REFERENCES chats(id_chat) ON DELETE CASCADE,
    id_participant UUID NOT NULL REFERENCES participants(id_participant),
    id_chats_folder UUID NOT NULL REFERENCES chats_folders(id_chats_folder) ON DELETE CASCADE,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(id_chat, id_participant, id_chats_folder)
);

-- Индексы для ускорения запросов
CREATE INDEX IF NOT EXISTS idx_chat_senders_chat ON chat_senders(id_chat);
CREATE INDEX IF NOT EXISTS idx_chat_senders_participant ON chat_senders(id_participant);
CREATE INDEX IF NOT EXISTS idx_messages_chat ON messages(id_chat);
CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(id_sender);
CREATE INDEX IF NOT EXISTS idx_chats_in_folder_chat ON chats_in_folder(id_chat);
CREATE INDEX IF NOT EXISTS idx_chats_in_folder_folder ON chats_in_folder(id_chats_folder);
CREATE INDEX IF NOT EXISTS idx_archive_chat_participant ON archive_chat(id_participant);
CREATE INDEX IF NOT EXISTS idx_pinned_chat_participant ON pinned_chat(id_participant);
`
