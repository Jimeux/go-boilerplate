create table if not exists model (
  id         int auto_increment primary key,
  name       varchar(255) not null,
  created_at timestamp    not null default current_timestamp,
  updated_at timestamp    not null default current_timestamp,
  deleted_at timestamp    null     default null
);

insert into model (name)
values ('Name 1'),
       ('Name 2'),
       ('Name 3'),
       ('Name 4'),
       ('Name 5'),
       ('Name 6');
