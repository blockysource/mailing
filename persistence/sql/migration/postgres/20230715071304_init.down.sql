BEGIN;

DROP TABLE mailing_message_queue;
DROP TABLE mailing_message_parameter;
DROP TABLE mailing_message_attachment;
DROP TABLE mailing_message_bcc_address;
DROP TABLE mailing_message_cc_address;
DROP TABLE mailing_message_to_address;
DROP TABLE mailing_message;
DROP TABLE mailing_template_parameter;
DROP TABLE mailing_template;
DROP TABLE mailing_provider;
DROP TABLE mailing_provider_type;

COMMIT;