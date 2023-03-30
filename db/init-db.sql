SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;
SET search_path = public, pg_catalog;

CREATE UNLOGGED TABLE IF NOT EXISTS actor (
    nickname VARCHAR(100) primary key,
    fullname VARCHAR(400) NOT NULl,
    about TEXT NOT NULl default '',
    email VARCHAR(150)
);

CREATE UNIQUE INDEX if not exists test_email on actor (lower(email));
CREATE UNIQUE INDEX if not exists test_nickname on actor (lower(nickname));

CREATE UNLOGGED TABLE IF NOT EXISTS forum(
    slug VARCHAR(100) primary key,
    title VARCHAR(100) NOT NULL,
    host VARCHAR(100) NOT NULL references actor (nickname) on DELETE CASCADE,
    posts bigint default 0,
    threads bigint default 0
);

CREATE UNIQUE INDEX if not exists test_forum on forum (lower(slug));

CREATE SEQUENCE IF NOT EXISTS thread_id_seq;

CREATE UNLOGGED TABLE IF NOT EXISTS thread(
    id bigint primary key default nextval('thread_id_seq'),
    title VARCHAR(300) NOT NULL,
    author VARCHAR(100) NOT NULL references actor (nickname) on DELETE CASCADE,
    forum VARCHAR(100) references forum (slug) on DELETE CASCADE,
    message TEXT NOT NULL,
    slug VARCHAR(150),
    created timestamp with time zone DEFAULT now(),
    votes int default 0
);

CREATE UNIQUE INDEX if not exists test_thread on thread (lower(slug));
CREATE INDEX if not exists thread_lower_forum on thread (lower(forum));

CREATE SEQUENCE IF NOT EXISTS post_id_seq;

CREATE UNLOGGED TABLE IF NOT EXISTS post(
    id bigint primary key default nextval('post_id_seq'),
    parent bigint,
    author VARCHAR(100) references actor(nickname) on DELETE CASCADE,
    message TEXT NOT NULL,
    isEdited boolean,
    forum VARCHAR(100) references forum(slug) on DELETE CASCADE not null,
    threadid bigint references thread(id) on DELETE CASCADE not null,
    created timestamp DEFAULT now(),
    pathtree bigint[]  default array []::bigint[]
);

CREATE SEQUENCE IF NOT EXISTS vote_id_seq;

CREATE UNLOGGED TABLE IF NOT EXISTS vote(
    id bigint primary key default nextval('vote_id_seq'),
    threadid bigint references thread(id) on DELETE CASCADE not null,
    nickname VARCHAR(100) references actor(nickname) on DELETE CASCADE not null,
    voice smallint not null,
    constraint unique_voice unique(threadid,nickname)
);

CREATE UNLOGGED TABLE IF NOT EXISTS forum_actors(
    nickname VARCHAR(100) references actor(nickname) on DELETE CASCADE not null,
    fullname VARCHAR(400) not null,
    about TEXT not null default '',
    email VARCHAR(150) ,
    forum VARCHAR(100) references forum (slug)
);

CREATE UNIQUE INDEX if not exists uni_actor_in_forum on forum_actors (lower(nickname),lower(forum));

DO $$ BEGIN
    CREATE TYPE usmeta AS (
                              fullname VARCHAR(400),
                              about text,
                              email VARCHAR(150)
                          );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE OR REPLACE FUNCTION insertForumActors() RETURNS trigger as
$insertForumActors$
begin
    insert into forum_actors (nickname, fullname, about, email, forum)
    values (new.author,'usermeta.fullname','usermeta.about','usermeta.email',new.forum)
    on conflict do nothing;
    return new;
end;
$insertForumActors$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION insertPathTree() RETURNS trigger as
$insertPathTree$
Declare
    parent_path         BIGINT[];
begin
    if ( new.parent = 0 ) then
        new.pathtree := array_append(new.pathtree,new.id);
    else
        select pathtree from post where id = new.parent into parent_path;
    new.pathtree := new.pathtree || parent_path || new.id;
    end if;
    UPDATE forum SET posts=posts + 1 WHERE lower(forum.slug) = lower(new.forum);
    return new;
end;
$insertPathTree$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION addPostsToForum() RETURNS trigger as
$addPostsToForum$
BEGIN
    UPDATE forum SET posts=posts + 1 WHERE lower(forum.slug) = lower(new.forum);
    return new;
end;
$addPostsToForum$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION addThreadsToForum() RETURNS trigger as
$addThreadsToForum$
BEGIN
    UPDATE forum SET threads=threads + 1 WHERE lower(forum.slug) = lower(new.forum);
    return new;
end;
$addThreadsToForum$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION insertThreadsVotes() RETURNS trigger as
$insertThreadsVotes$
begin
    update thread set votes = votes + new.voice where id = new.threadid;
    return new;
end;
$insertThreadsVotes$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION updateThreadsVotes() RETURNS trigger as
$updateThreadsVotes$
begin
    update thread set votes = votes + new.voice - old.voice where id = new.threadid;
    return new;
end;
$updateThreadsVotes$ LANGUAGE plpgsql;

drop trigger if exists insertPathTreeTrigger on post;
drop trigger if exists insertThreadsVotesTrigger on vote;
drop trigger if exists updateThreadsVotesTrigger on vote;
drop trigger if exists insertForumActorsThreadTrigger on thread;
drop trigger if exists insertForumActorsPostTrigger on post;
drop trigger if exists updateThreadsForumCount on thread;
drop trigger if exists updatePostsForumCount on post;

CREATE TRIGGER  insertPathTreeTrigger
    BEFORE INSERT
    on post
    for each row
EXECUTE Function insertPathTree();

CREATE TRIGGER  insertThreadsVotesTrigger
    AFTER INSERT
    on vote
    for each row
EXECUTE Function insertThreadsVotes();

CREATE TRIGGER  updateThreadsVotesTrigger
    AFTER UPDATE
    on vote
    for each row
EXECUTE Function updateThreadsVotes();

CREATE TRIGGER  insertForumActorsThreadTrigger
    AFTER INSERT
    on thread
    for each row
EXECUTE Function insertForumActors();

CREATE TRIGGER  insertForumActorsPostTrigger
    AFTER INSERT
    on post
    for each row
EXECUTE Function insertForumActors();

CREATE TRIGGER  updateThreadsForumCount
    AFTER INSERT
    on thread
    for each row
EXECUTE Function addThreadsToForum();

CREATE INDEX IF NOT EXISTS forum_slug_hash ON forum using hash (slug);

CREATE INDEX IF NOT EXISTS thread_slug ON thread using hash (lower(slug));
CREATE INDEX IF NOT EXISTS thread_id_forum ON thread using hash (lower(forum));
CREATE INDEX IF NOT EXISTS thread_slug_id ON thread (lower(slug), id);
CREATE INDEX IF NOT EXISTS actor_ascii_nickname on forum_actors using hash (lower(nickname) collate "C");
CREATE INDEX IF NOT EXISTS vote_nickname ON vote (lower(nickname), threadid, voice);

CREATE INDEX IF NOT EXISTS post_path ON post using hash((pathtree));
CREATE INDEX IF NOT EXISTS post_path ON post (id,(pathtree));
CREATE INDEX IF NOT EXISTS post_threadid ON post (threadid, id);

CREATE INDEX IF NOT EXISTS post_thread_parent_id_threadid ON post (threadid, parent, id); -- parent tree sort
CREATE INDEX IF NOT EXISTS post_first_parent_id ON post ((pathtree[1]),id); -- parent tree sort
CREATE INDEX IF NOT EXISTS thread_parenttree_post on post (threadid,pathtree); -- tree sort
CREATE INDEX IF NOT EXISTS post_first_parent_thread ON post ((pathtree[1]), threadid); -- tree sort
CREATE INDEX IF NOT EXISTS forum_actor_forum ON forum_actors using hash (lower(forum));
