drop table if exists users cascade;
drop table if exists surveys cascade;
drop table if exists questions cascade;

create table users (
    id serial primary key,
    email varchar(255) not null unique,
    password varchar(255) not null
);

create table surveys (
    id serial primary key,
    user_id int not null,
    title varchar(255) not null,
    created timestamptz,
    updated timestamptz,
    questions_order int[] default('{}') not null,
    constraint fk_user foreign key(user_id) references users(id) on delete cascade
);

create table questions (
    id serial primary key,
    user_id int not null,
    survey_id int not null,
    title varchar(255) default('') not null,
    description text default('') not null,
    options jsonb default('[]'::json) not null,
    constraint fk_survey foreign key(survey_id) references surveys(id) on delete cascade,
    constraint fk_user foreign key(user_id) references users(id) on delete cascade
);

create index question_survey_idx on questions (survey_id);
