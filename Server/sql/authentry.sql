-- Table: public.authentry

-- DROP TABLE public.authentry;

CREATE TABLE public.authentry
(
    userid character varying(320) COLLATE pg_catalog."default" NOT NULL,
    token character varying(64) COLLATE pg_catalog."default" NOT NULL,
    ts timestamp without time zone NOT NULL DEFAULT now(),
    CONSTRAINT authentry_pkey PRIMARY KEY (userid, token)
)

TABLESPACE pg_default;

ALTER TABLE public.authentry OWNER to iagd_lambda;

COMMENT ON TABLE public.authentry
    IS 'GDIA: Auth tokens for the backup API';