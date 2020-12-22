-- Table: public.authattempt

-- DROP TABLE public.authattempt;

CREATE TABLE public.authattempt
(
    key character varying(36) COLLATE pg_catalog."default" NOT NULL,
    code character varying(9) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    userid character varying(340) COLLATE pg_catalog."default",
    CONSTRAINT authattempt_pkey PRIMARY KEY (key, code)
)

TABLESPACE pg_default;

ALTER TABLE public.authattempt OWNER to iagd_lambda;

COMMENT ON TABLE public.authattempt
    IS 'Contains a publicly known "token" and a secret pin code used to authenticate for a given user. 

Upon presenting both the token and the code to an API, an access token is inserted into "authentry" and returned to the user.';