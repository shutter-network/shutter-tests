--- find decryption key by preimage

select * from decryption_key where identity_preimage=decode('75c3b90000000000000000000000000000000000000000000000000000000000f1fc0e5b6c5e42639d27ab4f2860e964de159bb4', 'hex');

--- find decryption_keys_message by preimage

select decryption_keys_message_slot from decryption_key as dk left join decryption_keys_message_decryption_key as dm on dk.id=dm.decryption_key_id where identity_preimage=decode('75c3b90000000000000000000000000000000000000000000000000000000000f1fc0e5b6c5e42639d27ab4f2860e964de159bb4', 'hex');

--- find tx for preimage

--- calculate tester ratios from observer

SELECT
        COUNT(*) AS known_tx, 
        SUM(CASE WHEN dt.tx_status='shielded inclusion' THEN 1.0 END)/COUNT(*) * 100 AS shielded_ratio,    
        SUM(CASE WHEN dt.tx_status='shielded inclusion' THEN 1.0 END) AS shielded_amount,    
        SUM(CASE WHEN dt.tx_status='unshielded inclusion' THEN 1.0 END)/COUNT(*) * 100 AS unshielded_ratio,
        SUM(CASE WHEN dt.tx_status='unshielded inclusion' THEN 1.0 END) AS unshielded_amount,
        SUM(CASE WHEN dt.tx_status='not included' THEN 1.0 END)/COUNT(*) * 100 AS not_included_ratio, 
        SUM(CASE WHEN dt.tx_status='not included' THEN 1.0 END) not_included_amount, 
        SUM(CASE WHEN dt.tx_status='pending' THEN 1.0 END)/COUNT(*) * 100 AS missing_key_ratio,
        SUM(CASE WHEN dt.tx_status='pending' THEN 1.0 END) AS missing_key_amount
        FROM decryption_key AS dk 
                LEFT JOIN decrypted_tx AS dt
                        ON dt.decryption_key_id=dk.id
                LEFT JOIN block AS b
                        ON b.slot=dt.slot 
WHERE 
        SUBSTRING(
                ENCODE(dk.identity_preimage, 'hex'),  --- encode preimage as hex string
                65  --- match only sender suffix of identity_preimage
        ) = '0ce7efc29fffbe1cd3bd13b4cbdc0ae07def3a90'  --- address of tester account
AND 
        b.block_number BETWEEN 38389988 AND 38405492;



--- convert identity preimage to block numbers

--- 1) add function to swap from little endian to big endian (SO DBA a/337446)

CREATE OR REPLACE FUNCTION bytea_reverse(bytea)
RETURNS bytea                                                                                           language sql immutable PARALLEL safe
RETURN (          
    SELECT string_agg(substring($1 FROM byte_pos FOR 1), '')
      FROM generate_series(octet_length($1), 1, -1) AS byte_pos);

--- 2) create function for prefix to block number

CREATE OR REPLACE FUNCTION prefix_to_blocknumber(bytea)
RETURNS numeric
        language sql immutable PARALLEL safe
RETURN (
    SELECT ('0x' || ENCODE(bytea_reverse(SUBSTRING($1 FROM 0 FOR 32)), 'hex'))::float::numeric);

--- 3) extract block number identity prefix

SELECT
        ('0x' || ENCODE(bytea_reverse(SUBSTRING(dk.identity_preimage FROM 0 FOR 32)), 'hex'))::float::numeric,
        prefix_to_blocknumber(dk.identity_preimage)
        FROM decryption_key AS dk
                LEFT JOIN decrypted_tx AS dt
                        ON dt.decryption_key_id=dk.id
                LEFT JOIN block AS b
                        ON b.slot=dt.slot
WHERE
        SUBSTRING(
                ENCODE(dk.identity_preimage, 'hex'),  --- encode preimage as hex string
                65  --- match only sender suffix of identity_preimage
        ) = '0ce7efc29fffbe1cd3bd13b4cbdc0ae07def3a90'  --- address of tester account
AND
        b.block_number BETWEEN 38389988 AND 38405492 AND NOT dt.tx_status='not included';

--- find late sequenced tx

SELECT 
    e.event_block_number, 
    prefix_to_blocknumber(e.identity_prefix) 
FROM transaction_submitted_event AS e 
    LEFT JOIN decryption_key AS dk 
    ON dk.identity_preimage=(e.identity_prefix||e.sender) 
WHERE 
    event_block_number BETWEEN 38389988 AND 38406335 
    AND sender='\x0ce7efc29fffbe1cd3bd13b4cbdc0ae07def3a90'::bytea 
    AND NOT e.event_block_number=(prefix_to_blocknumber(e.identity_prefix)+1);



--- inspect triggers

SELECT
    (e.block_number - b.block_number), 
    e.identity_prefix 
FROM proposer_duties AS p 
    LEFT JOIN validator_status AS v 
        ON v.validator_index=p.validator_index 
    LEFT JOIN block AS b 
        ON b.slot=p.slot 
    LEFT JOIN transaction_submitted_event AS e 
        ON prefix_to_blocknumber(e.identity_prefix)=b.block_number 
WHERE 
    b.block_number BETWEEN 38389989 AND 38406404 
    AND v.status='active_ongoing' 
    AND e.sender='\x0ce7efc29fffbe1cd3bd13b4cbdc0ae07def3a90'::bytea;

---  inspect sequenced tx

SELECT 
    COUNT(e.event_block_number) 
FROM proposer_duties AS p 
    LEFT JOIN validator_status AS v 
        ON v.validator_index=p.validator_index 
    LEFT JOIN block AS b 
        ON b.slot=p.slot 
    LEFT JOIN transaction_submitted_event AS e 
        ON prefix_to_blocknumber(e.identity_prefix)=b.block_number 
WHERE 
    b.block_number BETWEEN 38389988 AND 38406946 
    AND v.status='active_ongoing' 
    AND e.sender='\x0ce7efc29fffbe1cd3bd13b4cbdc0ae07def3a90'::bytea;

--- count decrypted test tx

SELECT 
    COUNT(d.id) 
FROM proposer_duties AS p 
    LEFT JOIN validator_status AS v 
        ON v.validator_index=p.validator_index 
    LEFT JOIN block AS b 
        ON b.slot=p.slot 
    LEFT JOIN transaction_submitted_event AS e 
        ON prefix_to_blocknumber(e.identity_prefix)=b.block_number 
    LEFT JOIN decryption_key AS d 
        ON d.identity_preimage=(e.identity_prefix || e.sender) 
WHERE 
    b.block_number BETWEEN 38389988 AND 38406946 
    AND v.status='active_ongoing' 
    AND e.sender='\x0ce7efc29fffbe1cd3bd13b4cbdc0ae07def3a90'::bytea;


--- count sequenced tx filtered by decrypted tx

SELECT count(*) 
FROM transaction_submitted_event AS e 
WHERE e.identity_prefix IN 
    (
        SELECT SUBSTRING(dk.identity_preimage FROM 0 FOR 33)::bytea AS prefix 
        FROM decrypted_tx AS dt 
        LEFT JOIN decryption_key AS dk 
        ON dk.id=dt.decryption_key_id
        WHERE prefix_to_blocknumber(dk.identity_preimage) BETWEEN 38389988 AND 38407446 
        AND         SUBSTRING(
                ENCODE(dk.identity_preimage, 'hex'),  --- encode preimage as hex string
                65  --- match only sender suffix of identity_preimage
        ) = '0ce7efc29fffbe1cd3bd13b4cbdc0ae07def3a90'
);

--- count sequenced tx

SELECT count(*) 
FROM transaction_submitted_event AS e 
WHERE prefix_to_blocknumber(e.identity_prefix) BETWEEN 38389988 AND 38407446 
AND e.sender= '\x0ce7efc29fffbe1cd3bd13b4cbdc0ae07def3a90'::bytea;



