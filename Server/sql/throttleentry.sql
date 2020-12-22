CREATE TABLE public.throttleentry
(
    id bigserial NOT NULL,
    userid character varying(320),
    ip character varying(512),
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    CONSTRAINT throttle_pk PRIMARY KEY (id)
);

ALTER TABLE public.throttleentry
    OWNER to iagd_lambda;

COMMENT ON TABLE public.throttleentry
    IS 'GDIA: Throttle entries to prevent brute force attempts / email spam';