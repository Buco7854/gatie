-- +goose Up

-- Roles (reference table)
CREATE TABLE roles (
    id          VARCHAR(20) PRIMARY KEY,
    description TEXT NOT NULL
);

-- Permissions (reference table)
CREATE TABLE permissions (
    id          VARCHAR(50) PRIMARY KEY,
    description TEXT NOT NULL
);

-- Role-permission associations
CREATE TABLE role_permissions (
    role_id       VARCHAR(20) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id VARCHAR(50) NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Seed roles
INSERT INTO roles (id, description) VALUES
    ('ADMIN',   'Full access — bypasses all permission checks'),
    ('MANAGER', 'Can open/close gates, configure them, and manage gate members'),
    ('MEMBER',  'Can open/close gates and view status'),
    ('VIEWER',  'Can only view gate status');

-- Seed permissions
INSERT INTO permissions (id, description) VALUES
    ('*',                   'Bypass all permission checks'),
    ('gate:open',           'Open a gate'),
    ('gate:close',          'Close a gate'),
    ('gate:view_status',    'View gate status'),
    ('gate:configure',      'Configure gate settings'),
    ('gate:manage_members', 'Manage gate member access');

-- Seed role-permission associations
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('ADMIN',   '*'),
    ('MANAGER', 'gate:open'),
    ('MANAGER', 'gate:close'),
    ('MANAGER', 'gate:view_status'),
    ('MANAGER', 'gate:configure'),
    ('MANAGER', 'gate:manage_members'),
    ('MEMBER',  'gate:open'),
    ('MEMBER',  'gate:close'),
    ('MEMBER',  'gate:view_status'),
    ('VIEWER',  'gate:view_status');

-- Members
CREATE TABLE members (
    id            UUID PRIMARY KEY DEFAULT uuidv7(),
    username      VARCHAR(100) UNIQUE NOT NULL,
    display_name  VARCHAR(200),
    password_hash TEXT NOT NULL,
    role_id       VARCHAR(20) NOT NULL DEFAULT 'MEMBER' REFERENCES roles(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE members;
DROP TABLE role_permissions;
DROP TABLE permissions;
DROP TABLE roles;
