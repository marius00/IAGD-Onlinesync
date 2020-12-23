CREATE TABLE public.users
(
    userid character varying(320) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    buddy_id integer,
    PRIMARY KEY (userid)
);

ALTER TABLE public.users OWNER to iagd_lambda;

COMMENT ON COLUMN public.users.buddy_id
    IS 'Activated/created when the user enabled buddy sharing in settings in IA';
	
COMMENT ON TABLE public.users
  IS 'List of users in the backup system.
Helps keep track of new users and returning users (check if they have items in the old solution, notify them that they may have entered the wrong email etc)';

ALTER TABLE public.users ADD CONSTRAINT users_uq_buddy_id UNIQUE (buddy_id);