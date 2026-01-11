-- Sample SQL queries for testing batch analysis
-- This file contains multiple queries to demonstrate the batch feature

-- Query 1: Simple SELECT
SELECT * FROM users WHERE age > 25;

-- Query 2: JOIN operation
SELECT o.id, o.order_date, u.name
FROM orders o
JOIN users u ON o.user_id = u.id
WHERE o.status = 'pending';

-- Query 3: Aggregate query
SELECT user_id, COUNT(*) as order_count, SUM(total_amount) as total_spent
FROM orders
WHERE order_date >= '2024-01-01'
GROUP BY user_id
HAVING COUNT(*) > 5;