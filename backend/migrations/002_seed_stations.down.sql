-- Remove seeded stations
DELETE FROM stations WHERE crs_code IN (
    'WAT', 'VIC', 'PAD', 'EUS', 'KGX', 'STP', 'LST', 'CHX', 'LBG',
    'KNG', 'SUR', 'WIM', 'CLJ', 'RMD', 'EPC', 'BTN', 'GAT', 'ACT', 'EAL',
    'HHE', 'OLY', 'STF', 'MAN', 'BHM', 'LDS', 'NCL', 'EDB', 'GLC',
    'BRI', 'OXF', 'CAM', 'RDG'
);
