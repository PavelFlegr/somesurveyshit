drop table if exists surveys cascade;
drop table if exists questions cascade;

create table surveys (
    id serial primary key,
    title varchar(255) not null,
    created timestamptz,
    updated timestamptz,
    questions_order int[] default('{}') not null
);

create table questions (
    id serial primary key,
    survey_id int not null,
    title varchar(255) default('') not null,
    description text default('') not null,
    options jsonb default('[]'::json) not null,
    constraint fk_survey foreign key(survey_id) references surveys(id) on delete cascade
);

create index question_survey_idx on questions (survey_id);
