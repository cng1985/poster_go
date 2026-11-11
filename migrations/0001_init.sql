create table account (
    id integer primary key autoincrement,
    username text not null,
    password_hash text not null,
    display_name text not null,
    status text not null,
    created_at datetime not null,
    updated_at datetime not null
);

create unique index ux_account_username on account(username);

create table session (
    id integer primary key autoincrement,
    account_id integer not null references account(id),
    token_hash text not null,
    expires_at datetime not null,
    revoked_at datetime,
    created_at datetime not null
);

create unique index ux_session_token_hash on session(token_hash);
create index ix_session_account on session(account_id);

create table poster_template (
    id integer primary key autoincrement,
    code text not null,
    name text not null,
    category text not null,
    summary text not null,
    width integer not null,
    height integer not null,
    cover_color text not null,
    document text not null,
    sort_no integer not null,
    status text not null,
    created_at datetime not null,
    updated_at datetime not null
);

create unique index ux_poster_template_code on poster_template(code);
create index ix_poster_template_category on poster_template(category, sort_no, id);

create table poster (
    id integer primary key autoincrement,
    owner_id integer not null references account(id),
    title text not null,
    width integer not null,
    height integer not null,
    cover_color text not null,
    document text not null,
    version integer not null,
    created_at datetime not null,
    updated_at datetime not null
);

create index ix_poster_owner_updated on poster(owner_id, updated_at, id);

create table poster_generation (
    id integer primary key autoincrement,
    owner_id integer not null references account(id),
    poster_id integer,
    prompt text not null,
    style text not null,
    size_preset text not null,
    provider text not null,
    model text not null,
    document text not null,
    message text not null,
    created_at datetime not null
);

create index ix_poster_generation_owner on poster_generation(owner_id, id);

create table poster_analysis (
    id integer primary key autoincrement,
    owner_id integer not null references account(id),
    poster_id integer not null,
    score integer not null,
    summary text not null,
    suggestions text not null,
    provider text not null,
    created_at datetime not null
);

create index ix_poster_analysis_owner on poster_analysis(owner_id, id);
create index ix_poster_analysis_poster on poster_analysis(poster_id, id);

create table asset (
    id integer primary key autoincrement,
    owner_id integer not null references account(id),
    storage_key text not null,
    original_name text not null,
    content_type text not null,
    byte_size integer not null,
    width integer not null,
    height integer not null,
    created_at datetime not null
);

create unique index ux_asset_storage_key on asset(storage_key);
create index ix_asset_owner on asset(owner_id, id);
