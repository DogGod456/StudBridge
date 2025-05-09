package database

const Schema = `
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Типы участников
CREATE TYPE message_status AS ENUM ('sent', 'delivered', 'read', 'failed');

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
    ref_id TEXT NOT NULL,
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
    id_participant UUID NOT NULL REFERENCES participants(id_participant),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    UNIQUE(id_chat, id_participant)
);

-- Таблица сообщений
CREATE TABLE IF NOT EXISTS messages (
    id_message UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL REFERENCES chats(id_chat) ON DELETE CASCADE,
    id_sender UUID NOT NULL REFERENCES participants(id_participant),
    message_text TEXT NOT NULL,
    status message_status NOT NULL DEFAULT 'sent',
    id_parent_message UUID,
    draft BOOLEAN NOT NULL DEFAULT FALSE,
    sending_time TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);


-- Дополнительные таблицы
-- Таблица архива
CREATE TABLE IF NOT EXISTS archive_chat (
    id_archive_chat UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL,
    id_participant UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (id_chat) REFERENCES chats(id_chat) ON DELETE CASCADE,
    FOREIGN KEY (id_participant) REFERENCES participants(id_participant) ON DELETE CASCADE
);

-- Таблица папки
CREATE TABLE IF NOT EXISTS chats_folder (
    id_chats_folder UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_participant UUID NOT NULL,
    folder_name VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (id_participant) REFERENCES participants(id_participant) ON DELETE CASCADE

);

-- Таблица чатов в папке
CREATE TABLE IF NOT EXISTS chats_in_folder (
    id_chats_in_folder UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL,
    id_chats_folder UUID NOT NULL,
    FOREIGN KEY (id_chat) REFERENCES chats(id_chat) ON DELETE CASCADE,
    FOREIGN KEY (id_chats_folder) REFERENCES chats_folder(id_chats_folder) ON DELETE CASCADE
);

-- Таблица закреплённых чатов
CREATE TABLE IF NOT EXISTS pinned_chat (
    id_pinned_chat UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    id_chat UUID NOT NULL,
    id_participant UUID NOT NULL,
    id_chats_folder UUID NOT NULL,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (id_chat) REFERENCES chats(id_chat) ON DELETE CASCADE,
    FOREIGN KEY (id_participant) REFERENCES participants(id_participant) ON DELETE CASCADE,
    FOREIGN KEY (id_chats_folder) REFERENCES chats_folder(id_chats_folder) ON DELETE CASCADE

);

`
