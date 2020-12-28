-- Table: public.item

-- DROP TABLE public.item;

CREATE TABLE public.item
(
    id character varying(36) COLLATE pg_catalog."default" NOT NULL,
    userid character varying(320) COLLATE pg_catalog."default" NOT NULL,
    baserecord character varying(255) COLLATE pg_catalog."default" NOT NULL,
    prefixrecord character varying(255) COLLATE pg_catalog."default",
    suffixrecord character varying(255) COLLATE pg_catalog."default",
    modifierrecord character varying(255) COLLATE pg_catalog."default",
    transmuterecord character varying(255) COLLATE pg_catalog."default",
    seed bigint NOT NULL,
    materiarecord character varying(255) COLLATE pg_catalog."default",
    reliccompletionbonusrecord character varying(255) COLLATE pg_catalog."default",
    relicseed bigint,
    enchantmentrecord character varying(255) COLLATE pg_catalog."default",
    prefixrarity bigint,
    unknown bigint,
    enchantmentseed bigint,
    materiacombines bigint,
    stackcount bigint NOT NULL,
    name character varying(255) COLLATE pg_catalog."default",
    namelowercase character varying(255) COLLATE pg_catalog."default",
    rarity character varying(255) COLLATE pg_catalog."default",
    levelrequirement double precision,
    mod character varying(255) COLLATE pg_catalog."default",
    ishardcore boolean,
    created_at bigint,
    ts bigint NOT NULL,
    CONSTRAINT item_pkey PRIMARY KEY (id, userid)
)

TABLESPACE pg_default;

ALTER TABLE public.item OWNER to iagd_lambda;

COMMENT ON TABLE public.item
    IS 'GDIA: Items for the backup system';

COMMENT ON COLUMN public.item.id
    IS 'GUID provided by client';

COMMENT ON COLUMN public.item.userid
    IS 'User email address';
-- Index: idx_item_userid_ts

-- DROP INDEX public.idx_item_userid_ts;

CREATE INDEX idx_item_userid_ts
    ON public.item USING btree
    (userid COLLATE pg_catalog."default" ASC NULLS LAST, ts DESC NULLS LAST)
    TABLESPACE pg_default;

ALTER TABLE public.item
    CLUSTER ON idx_item_userid_ts;
	
ALTER TABLE public.item ALTER COLUMN levelrequirement TYPE integer;