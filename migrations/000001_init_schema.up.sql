create extension if not exists pg_trgm;

-- Shared trigger func
create or replace function auto_updated_at () returns TRIGGER as $$
begin
    NEW.updated_at = now();
    return NEW;
end;
$$ language 'plpgsql';

create type deck_visibility as enum('public', 'private');

create table deck (
    id uuid not null primary key default gen_random_uuid(),
    name varchar(255) NOT NULL,
    description varchar(1024),
    thumbnail text,
    visibility deck_visibility not null default 'private',
    archived_at timestamptz,
    created_at timestamptz not null DEFAULT now(),
    updated_at timestamptz
);

create index idx_deck_name_trgm on deck using gin (name gin_trgm_ops);

create index idx_deck_description_trgm on deck using gin (description gin_trgm_ops);

create trigger deck_updated_at before update on deck for each row execute function auto_updated_at();

create type deck_user_role as enum('member', 'admin');

create table deck_user (
    deck_id uuid references deck (id) on delete cascade,
    user_id text references auth."user" (id) on delete cascade,
    "role" deck_user_role not null default 'member',
    created_at timestamptz not null default now(),
    PRIMARY KEY (user_id, deck_id)
);

create index idx_deck_user_role on deck_user ("role");

create index idx_deck_user_user_id_deck_id on deck_user (user_id, deck_id);

create table card (
    id uuid not null primary key default gen_random_uuid(),
    deck_id uuid references deck (id) on delete cascade,
    name varchar(255) NOT NULL,
    description VARCHAR(1024),
    image text,
    archived_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

create index idx_card_deck_id on card(deck_id);

create index idx_card_name_trgm on card using gin (name gin_trgm_ops);

create trigger card_updated_at before update on card for each row execute function auto_updated_at();
