export const marketSnapshotSQL = `SELECT a.*,o.price::float8,o.market_cap::float8 cap,o.volume::float8 volume,o.change_24h::float8 change,o.observed_at updated,n.result analysis,
 COALESCE(h.history,'[]'::json) history FROM assets a
 LEFT JOIN LATERAL(SELECT * FROM observations WHERE asset_id=a.id ORDER BY observed_at DESC LIMIT 1)o ON true
 LEFT JOIN LATERAL(SELECT result FROM analytics WHERE asset_id=a.id ORDER BY day DESC LIMIT 1)n ON true
 LEFT JOIN LATERAL(SELECT json_agg(json_build_object('time',day::text,'price',price::float8) ORDER BY day) history FROM (SELECT day,price FROM daily_samples WHERE asset_id=a.id ORDER BY day DESC LIMIT 31)d)h ON true
 ORDER BY a.active DESC,o.market_cap DESC NULLS LAST,a.id LIMIT 250`;
