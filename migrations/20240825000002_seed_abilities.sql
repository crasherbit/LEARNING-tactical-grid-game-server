-- +goose Up
-- Seed abilities data
INSERT INTO abilities (id, name, base_damage, ap_cost, range, aoe_radius, per_turn_limit, per_target_per_turn_limit) VALUES 
('punch', 'Punch', 10, 2, 1, NULL, 3, NULL),
('fireball', 'Fireball', 20, 3, 4, 1, NULL, 1);

-- +goose Down
DELETE FROM abilities WHERE id IN ('punch', 'fireball');
