BEGIN;

-- mailing_template is a table that stores email templates.
CREATE TABLE mailing_template
(
    id           SERIAL PRIMARY KEY,
    uid          TEXT UNIQUE NOT NULL,
    from_address TEXT, -- from_address defines the default email address that will be used as the sender.
    name         TEXT        NOT NULL,
    subject      TEXT        NOT NULL,
    body         TEXT        NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- mailing_template_parameters is a table that stores predefined parameters for email templates.
CREATE TABLE mailing_template_parameter
(
    id            SERIAL PRIMARY KEY,
    template_id   INTEGER NOT NULL REFERENCES mailing_template (id),
    name          TEXT    NOT NULL,
    default_value TEXT,
    CONSTRAINT mailing_template_parameters_template_id_name_key
        UNIQUE (template_id, name)
);

-- template_parameters_template_id_idx is used to sort the email template parameters by template_id
CREATE INDEX mailing_template_parameters_template_id_idx
    ON mailing_template_parameter (template_id);


-- mailing_message is a table that stores the messages.
CREATE TABLE mailing_message
(
    id          BIGSERIAL PRIMARY KEY,
    uid         TEXT UNIQUE NOT NULL,
    template_id INTEGER     NOT NULL REFERENCES mailing_template (id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    body        TEXT        NOT NULL,
    subject     TEXT        NOT NULL
);

-- mailing_message_template_id_idx is used to sort the messages by template_id
CREATE INDEX mailing_message_template_id_idx
    ON mailing_message (template_id);

-- mailing_message_created_at_idx is used to sort the messages by creation date
CREATE INDEX mailing_message_created_at_idx
    ON mailing_message (created_at DESC);

-- mailing_message_to_address is a table that stores the to addresses for email messages.
CREATE TABLE mailing_message_to_address
(
    id         BIGSERIAL PRIMARY KEY,
    message_id BIGINT   NOT NULL,
    seq_num    SMALLINT NOT NULL,
    to_address TEXT     NOT NULL,
    CONSTRAINT mailing_message_to_address_message_id_fk
        FOREIGN KEY (message_id) REFERENCES mailing_message (id)
            ON DELETE CASCADE,
    CONSTRAINT mailing_message_to_address_message_id_seq_num_key
        UNIQUE (message_id, seq_num)
);

-- mailing_message_to_address_message_id_idx is used to sort the to addresses by message_id
CREATE INDEX mailing_message_to_address_message_id_idx
    ON mailing_message_to_address (message_id);

-- mailing_message_cc_address is a table that stores the cc addresses for email messages.
CREATE TABLE mailing_message_cc_address
(
    id         BIGSERIAL PRIMARY KEY,
    message_id BIGINT   NOT NULL,
    seq_num    SMALLINT NOT NULL,
    cc_address TEXT     NOT NULL,
    CONSTRAINT mailing_message_cc_address_message_id_fk
        FOREIGN KEY (message_id) REFERENCES mailing_message (id)
            ON DELETE CASCADE,
    CONSTRAINT mailing_message_cc_address_message_id_seq_num_key
        UNIQUE (message_id, seq_num)
);

-- mailing_message_cc_address_message_id_idx is used to sort the cc addresses by message_id
CREATE INDEX mailing_message_cc_address_message_id_idx
    ON mailing_message_cc_address (message_id);

-- mailing_message_bcc_address is a table that stores the bcc addresses for email messages.
CREATE TABLE mailing_message_bcc_address
(
    id          BIGSERIAL PRIMARY KEY,
    message_id  BIGINT   NOT NULL,
    seq_num     SMALLINT NOT NULL,
    bcc_address TEXT     NOT NULL,
    CONSTRAINT mailing_message_bcc_address_message_id_fk
        FOREIGN KEY (message_id) REFERENCES mailing_message (id)
            ON DELETE CASCADE,
    CONSTRAINT mailing_message_bcc_address_message_id_seq_num_key
        UNIQUE (message_id, seq_num)
);

-- mailing_message_bcc_address_message_id_idx is used to sort the bcc addresses by message_id
CREATE INDEX mailing_message_bcc_address_message_id_idx
    ON mailing_message_bcc_address (message_id);

-- mailing_message_attachment is a table that stores attachments for email messages.
CREATE TABLE mailing_message_attachment
(
    id           BIGSERIAL PRIMARY KEY,
    message_id   BIGINT   NOT NULL,
    seq_num      SMALLINT NOT NULL,
    filename     TEXT     NOT NULL,
    content_type TEXT     NOT NULL,
    filepath     TEXT     NOT NULL,
    ttl          BIGINT,
    CONSTRAINT mailing_message_attachment_message_id_fk
        FOREIGN KEY (message_id) REFERENCES mailing_message (id)
            ON DELETE CASCADE,
    CONSTRAINT mailing_message_attachment_message_id_seq_num_key
        UNIQUE (message_id, seq_num)
);

-- mailing_message_parameter is a table that stores provided parameters to be used in the email template
-- for given message.
CREATE TABLE mailing_message_parameter
(
    id                    BIGSERIAL PRIMARY KEY,
    message_id            BIGINT  NOT NULL,
    template_parameter_id INTEGER NOT NULL,
    value                 TEXT    NOT NULL,
    CONSTRAINT mailing_message_parameter_message_id_fk
        FOREIGN KEY (message_id) REFERENCES mailing_message (id)
            ON DELETE RESTRICT,
    CONSTRAINT mailing_message_parameter_message_id_fk
        FOREIGN KEY (message_id) REFERENCES mailing_message (id)
            ON DELETE CASCADE,
    CONSTRAINT mailing_message_parameter_message_id_template_parameter_id_key
        UNIQUE (message_id, template_parameter_id)
);

-- mailing_message_parameter_message_id_idx is an index on message_id column of mailing_message_parameter table.
CREATE INDEX mailing_message_parameter_message_id_idx
    ON mailing_message_parameter (message_id);

-- mailing_message_parameter_template_parameter_id_idx is an index on template_parameter_id column
-- of mailing_message_parameter table.
CREATE INDEX mailing_message_parameter_template_parameter_id_idx
    ON mailing_message_parameter (template_parameter_id);


-- mailing_message_queue is a table that is a queue of emails to be sent.
CREATE TABLE mailing_message_queue
(
    id          BIGSERIAL PRIMARY KEY,
    message_id  BIGINT      NOT NULL,
    enqueued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at     TIMESTAMPTZ,
    not_before  TIMESTAMPTZ,
    CONSTRAINT mailing_message_queue_message_id_fk
        FOREIGN KEY (message_id) REFERENCES mailing_message (id)
            ON DELETE CASCADE
);

-- mailing_message_queue_message_id_idx is an index on message_id column of mailing_message_queue table.
CREATE INDEX mailing_message_queue_message_id_idx
    ON mailing_message_queue (message_id);

-- mailing_message_queue_enqueued_at_idx is an index on enqueued_at column of mailing_message_queue table.
CREATE INDEX mailing_message_queue_enqueued_at_idx
    ON mailing_message_queue (enqueued_at DESC);

-- mailing_message_queue_sent_at_idx is an index on sent_at column of mailing_message_queue table.
CREATE INDEX mailing_message_queue_sent_at_idx
    ON mailing_message_queue (sent_at DESC NULLS FIRST);

-- mailing_message_queue_not_before_idx is used to sort the email queue by not_before date
CREATE INDEX mailing_message_queue_not_before_idx
    ON mailing_message_queue (not_before NULLS FIRST);

COMMIT;