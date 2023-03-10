-- name: get-full-municipality-keys
-- The parameter $1 is a regular expression which shall match the keys
SELECT DISTINCT key
FROM geodata.shapes
WHERE key ~ $1 AND length(key) = 12;

-- name: get-water-usages
-- The parameter $1 will be an array of municipal keys
SELECT year, sum(value) as usage
FROM water_usage.usages
WHERE municipal_key = ANY($1)
GROUP BY year
ORDER BY year;

-- name: get-current-population
SELECT year, sum(population) as pop
FROM population.current
WHERE municipality_key = ANY($1)
AND year >= $2::int
GROUP BY year
ORDER BY year;

-- name: get-future-population
SELECT year, sum(population) as pop
FROM population.prognosis
WHERE municipal_key = ANY($1)
AND migration_level = $2::migration_level
GROUP BY year
ORDER BY year;