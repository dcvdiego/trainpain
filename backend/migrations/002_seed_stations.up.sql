-- Seed common UK stations
-- In production, this would be populated from National Rail KnowledgeBase API

INSERT INTO stations (crs_code, station_name, latitude, longitude, station_type) VALUES
-- London terminals
('WAT', 'London Waterloo', 51.5031, -0.1132, 'national_rail'),
('VIC', 'London Victoria', 51.4952, -0.1438, 'national_rail'),
('PAD', 'London Paddington', 51.5154, -0.1755, 'national_rail'),
('EUS', 'London Euston', 51.5282, -0.1337, 'national_rail'),
('KGX', 'London Kings Cross', 51.5308, -0.1238, 'national_rail'),
('STP', 'London St Pancras International', 51.5321, -0.1260, 'national_rail'),
('LST', 'London Liverpool Street', 51.5179, -0.0817, 'national_rail'),
('CHX', 'London Charing Cross', 51.5081, -0.1247, 'national_rail'),
('LBG', 'London Bridge', 51.5049, -0.0863, 'national_rail'),

-- South West London
('KNG', 'Kingston', 51.4131, -0.3065, 'national_rail'),
('SUR', 'Surbiton', 51.3928, -0.3073, 'national_rail'),
('WIM', 'Wimbledon', 51.4214, -0.2064, 'national_rail'),
('CLJ', 'Clapham Junction', 51.4643, -0.1705, 'national_rail'),
('RMD', 'Richmond', 51.4613, -0.3013, 'national_rail'),

-- South London
('EPC', 'East Putney', 51.4586, -0.2115, 'national_rail'),
('BTN', 'Brighton', 50.8292, -0.1410, 'national_rail'),
('GAT', 'Gatwick Airport', 51.1567, -0.1615, 'national_rail'),

-- West London
('ACT', 'Acton Main Line', 51.5173, -0.2672, 'national_rail'),
('EAL', 'Ealing Broadway', 51.5153, -0.3017, 'national_rail'),

-- North London
('HHE', 'Highbury & Islington', 51.5462, -0.1038, 'national_rail'),
('OLY', 'Olympia', 51.4982, -0.2145, 'national_rail'),

-- East London
('STF', 'Stratford', 51.5416, 0.0042, 'national_rail'),

-- Major UK cities
('MAN', 'Manchester Piccadilly', 53.4777, -2.2309, 'national_rail'),
('BHM', 'Birmingham New Street', 52.4777, -1.8996, 'national_rail'),
('LDS', 'Leeds', 53.7954, -1.5488, 'national_rail'),
('NCL', 'Newcastle', 54.9683, -1.6177, 'national_rail'),
('EDB', 'Edinburgh Waverley', 55.9521, -3.1889, 'national_rail'),
('GLC', 'Glasgow Central', 55.8582, -4.2585, 'national_rail'),
('BRI', 'Bristol Temple Meads', 51.4490, -2.5813, 'national_rail'),
('OXF', 'Oxford', 51.7535, -1.2702, 'national_rail'),
('CAM', 'Cambridge', 52.1945, 0.1376, 'national_rail'),
('RDG', 'Reading', 51.4543, -0.9781, 'national_rail')
ON CONFLICT (crs_code) DO NOTHING;
