-- name: Get :one
SELECT *
FROM factcheck
WHERE id = $2
    AND EXISTS (
        select 1
        from factcheck_user
        where user_id = $1
            and factcheck_id = $2
    );
-- name: ListPermitted :many
SELECT fc.*
FROM factcheck fc
    INNER JOIN factcheck_user fu ON fc.id = fu.factcheck_id
WHERE fu.user_id = $1
    AND (
        sqlc.narg ('claim')::text IS NULL
        OR fc.claim ILIKE '%' || sqlc.narg ('claim') || '%'
    )
    AND (
        CASE
            sqlc.arg(archived)::boolean
            WHEN true THEN archived_at IS NOT NULL
            WHEN false THEN archived_at IS NULL
        END
    )
ORDER BY fc.created_at DESC,
    fc.id ASC
LIMIT $2 OFFSET $3;
-- name: TotalCount :one
SELECT count(fc.id) as total_count
FROM factcheck fc
    INNER JOIN factcheck_user fu ON fc.id = fu.factcheck_id
WHERE fu.user_id = $1
    AND (
        sqlc.narg ('claim')::text IS NULL
        OR fc.claim ILIKE '%' || sqlc.narg ('claim') || '%'
    )
    AND (
        CASE
            sqlc.arg(archived)::boolean
            WHEN true THEN archived_at IS NOT NULL
            WHEN false THEN archived_at IS NULL
        END
    );