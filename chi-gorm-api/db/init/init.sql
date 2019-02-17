create table if not exists model
(
  id   int auto_increment primary key,
  name varchar(255)
);

insert into model (name)
values ('Name 1'),
       ('Name 2'),
       ('Name 3'),
       ('Name 4'),
       ('Name 5'),
       ('Name 6');
