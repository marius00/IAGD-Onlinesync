-- Table: public.deleteditem

-- DROP TABLE public.deleteditem;

CREATE TABLE public.deleteditem
(
    userid character varying(320) COLLATE pg_catalog."default" NOT NULL,
    id character varying(36) COLLATE pg_catalog."default" NOT NULL,
    ts bigint NOT NULL,
    CONSTRAINT deleteditem_pkey PRIMARY KEY (userid, id)
)

TABLESPACE pg_default;

ALTER TABLE public.deleteditem OWNER to iagd_lambda;

COMMENT ON TABLE public.deleteditem
    IS 'GDIA: Items which have been deleted. ID is stored here so that other clients can sync down and delete the item.';
-- Index: idx_deleteditem_userid_ts

-- DROP INDEX public.idx_deleteditem_userid_ts;

CREATE INDEX idx_deleteditem_userid_ts
    ON public.deleteditem USING btree
    (userid COLLATE pg_catalog."default" ASC NULLS LAST, ts DESC NULLS LAST)
    TABLESPACE pg_default;