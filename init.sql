CREATE TABLE adverts (
    id SERIAL, 
    title VARCHAR(200), 
    ad_description VARCHAR(1000),
    price INT,
    photos text[3]
    );