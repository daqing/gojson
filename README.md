# gojson
An unix tool to parse JSON in shell scripts

# usage

  $ echo '{"name": "David", "age": 29, "level": 2.7}' | gojson name
  > David

  $ echo '{"name": "David", "age": 29, "level": 2.7}' | gojson age
  > 29

  $ echo '{"name": "David", "age": 29, "level": 2.7}' | gojson level
  > 2.700000
