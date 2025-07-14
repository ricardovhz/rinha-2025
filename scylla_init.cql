CREATE KEYSPACE IF NOT EXISTS paymentsks WITH REPLICATION = {'class': 'NetworkTopologyStrategy', 'replication_factor': 1};
CREATE TABLE IF NOT EXISTS paymentsks.payments (requested_at timestamp,id uuid,backend int,amount decimal,PRIMARY KEY (backend, requested_at, id));
DELETE from paymentsks.payments WHERE backend in (0,1);