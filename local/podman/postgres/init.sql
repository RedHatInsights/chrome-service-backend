--
ALTER user unleash_user superuser
    createdb
    createrole;
CREATE DATABASE unleash;
GRANT ALL PRIVILEGES ON DATABASE unleash to unleash_user;
GRANT ALL ON SCHEMA public to unleash_user;
GRANT ALL ON ALL TABLES IN SCHEMA public to unleash_user;
