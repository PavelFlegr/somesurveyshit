/*drop table if exists users cascade;*/
drop table if exists surveys cascade;
drop table if exists blocks cascade;
drop table if exists questions cascade;
drop table if exists survey_permissions cascade;
drop table if exists responses cascade;

create table if not exists users (
    id serial primary key,
    email varchar(255) not null unique,
    password varchar(255) not null
);

create table if not exists surveys (
    id serial primary key,
    user_id int not null,
    title varchar(255) not null,
    created timestamptz,
    updated timestamptz,
    blocks_order int[] default('{}') not null,
    constraint fk_user foreign key(user_id) references users(id) on delete cascade
);

create table if not exists blocks (
    id serial primary key,
    user_id int not null,
    survey_id int not null,
    title varchar(255) not null,
    randomize boolean default(false) not null,
    submit boolean default(false) not null,
    submit_after float default(1) not null,
    created timestamptz,
    questions_order int[] default('{}') not null,
    constraint fk_user foreign key(user_id) references users(id) on delete cascade,
    constraint fk_survey foreign key(survey_id) references surveys(id) on delete cascade

);

create table if not exists questions (
    id serial primary key,
    user_id int not null,
    survey_id int not null,
    block_id int not null,
    title varchar(255) default('') not null,
    description text default('') not null,
    configuration jsonb default('{}'::json) not null,
    constraint fk_user foreign key(user_id) references users(id) on delete cascade,
    constraint fk_survey foreign key(survey_id) references surveys(id) on delete cascade,
    constraint fk_block foreign key(block_id) references blocks(id) on delete cascade
);

create index question_survey_idx on questions (survey_id);

create table if not exists survey_permissions (
    user_id int not null,
    action varchar(255) not null,
    entity_id int not null,
    constraint fk_user foreign key(user_id) references users(id) on delete cascade,
    constraint fk_survey foreign key(entity_id) references surveys(id) on delete cascade,
    primary key (user_id, action, entity_id)
);

create table if not exists responses (
    id serial primary key,
    survey_id int not null,
    response jsonb not null,
    constraint fk_survey foreign key (survey_id) references surveys(id) on delete cascade
);

create index response_survey_idx on responses (survey_id);