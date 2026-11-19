-- =============================================================================
-- AegisOpsAI — PostgreSQL Initialization
-- =============================================================================
-- This script runs once when the PostgreSQL container is first created.

-- Enable pgvector for similarity search on embeddings
CREATE EXTENSION IF NOT EXISTS vector;

-- Enable uuid-ossp for UUID generation functions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
