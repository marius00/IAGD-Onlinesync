-- Table: public.character

-- DROP TABLE public."character";

CREATE TABLE public."character"
(
    userid character varying(320) COLLATE pg_catalog."default" NOT NULL,
    name character varying(50) COLLATE pg_catalog."default" NOT NULL,
    filename character varying(400) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    CONSTRAINT character_pkey PRIMARY KEY (userid, name)
)

TABLESPACE pg_default;

ALTER TABLE public."character" OWNER to iagd_lambda;

COMMENT ON TABLE public."character" IS 'Stores filename mappings for character backups';

ALTER TABLE public."character" ADD COLUMN "updated_at" timestamp without time zone NOT NULL DEFAULT now();