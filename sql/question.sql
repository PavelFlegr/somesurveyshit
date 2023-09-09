drop table if exists questions cascade;
drop table if exists options;

create table questions (
    id serial primary key,
    title varchar(255) default('') not null,
    description text default('') not null,
    options_order int[] default('{}') not null
);

create table options (
    id serial primary key,
    question_id int,
    value varchar(255) default('') not null,
    constraint fk_question
        foreign key(question_id) references questions(id) on delete cascade
);

create index option_question_idx on options (question_id)