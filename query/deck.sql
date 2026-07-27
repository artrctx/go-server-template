-- name: Get :one
select *
from deck
where id = $2
    and exists (
        select 1
        from deck_user
        where user_id = $1
            and deck_id = $2
    );
-- name: ListPermitted :many
select d.*
from deck d
    inner join deck_user du on d.id = du.deck_id
where du.user_id = $1
    and (
        sqlc.narg ('name')::text is null
        or d.name ILIKE '%' || sqlc.narg ('name') || '%'
    )
    and (
        sqlc.narg ('description')::text is null
        or d.description ILIKE '%' || sqlc.narg ('description') || '%'
    )
    and (
        sqlc.narg ('archived')::boolean is true
        or d.archived_at is null
    )
    and (
        sqlc.narg ('visibility')::text is null
        or d.visibility::text = sqlc.narg ('visibility')
    )
order by d.created_at desc,
    d.id asc
limit $2 offset $3;
-- name: PermittedCount :one
select count(d.id) as total_count
from deck d
    inner join deck_user du on d.id = du.deck_id
where du.user_id = $1
    and (
        sqlc.narg ('name')::text is null
        or d.name ILIKE '%' || sqlc.narg ('name') || '%'
    )
    and (
        sqlc.narg ('description')::text is null
        or d.description ILIKE '%' || sqlc.narg ('description') || '%'
    )
    and (
        sqlc.narg ('archived')::boolean is true
        or d.archived_at is null
    )
    and (
        sqlc.narg ('visibility')::text is null
        or d.visibility::text = sqlc.narg ('visibility')
    );
-- name: FindPermitted :one
select d.*
from deck d
    left join deck_user du on d.id = du.deck_id
where d.id = $1
    and (
        d.visibility::text = 'public'
        or du.user_id = $2
    );
-- name: Create :one
insert into deck (name, description, thumbnail, visibility)
values($1, $2, $3, $4)
returning *;
-- name: LinkUser :exec
insert into deck_user(deck_id, user_id, role)
values($1, $2, 'admin');
-- name: Update :one
update deck d
set name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    thumbnail = COALESCE(sqlc.narg('thumbnail'), thumbnail),
    visibility = COALESCE(
        sqlc.narg('visibility'),
        visibility
    ),
    updated_at = NOW()
FROM deck_user du
WHERE d.id = du.deck_id
    AND du.user_id = $1
    AND d.id = $2
returning d.*;
-- name: CheckReadPermission :one
select exists (
        select 1
        from deck_user du
            inner join deck d on d.id = du.deck_id
        where d.visibility = 'public'
            or (
                du.user_id = $1
                and du.deck_id = $2
            )
    ) as row_exists;
-- name: CheckWritePermission :one
select exists (
        select 1
        from deck_user du
        where du.user_id = $1
            and du.deck_id = $2
    ) as row_exists;