# Internal health check endpoints
location = /internal/health/location-producer {
    internal;
    proxy_pass http://location-producer:6969/health;
    proxy_set_header Host $host;
    proxy_connect_timeout 5s; 
    proxy_read_timeout 5s;     
    proxy_send_timeout 5s;     
}

location = /internal/health/matching-producer {
    internal;
    proxy_pass http://matching-producer:7979/health;
    proxy_set_header Host $host;
    proxy_connect_timeout 5s; 
    proxy_read_timeout 5s;     
    proxy_send_timeout 5s;     
}

location = /internal/health/map-service {
    internal;
    proxy_pass http://map-service:8080/health;
    proxy_set_header Host $host;
    proxy_connect_timeout 5s;
    proxy_read_timeout 5s;
    proxy_send_timeout 5s;
}

location = /internal/health/kafka {
    internal;
    proxy_pass http://kafka-ui:8080/actuator/health;
    proxy_set_header Host $host;
    proxy_connect_timeout 5s;
    proxy_read_timeout 5s;
    proxy_send_timeout 5s;
}

location /health {

    allow 127.0.0.1;        # localhost
    allow 10.0.0.0/8;       # Internal network
    allow 172.16.0.0/12;    # Docker network
    allow 192.168.0.0/16;  
    deny all; 

    default_type application/json;
    
    content_by_lua_block {
        local location_res = ngx.location.capture("/internal/health/location-producer")
        local matching_res = ngx.location.capture("/internal/health/matching-producer")
        local map_res = ngx.location.capture("/internal/health/map-service")
        local kafka_res = ngx.location.capture("/internal/health/kafka")
        
        local location_status = "down"
        local matching_status = "down"
        local map_status = "down"
        local kafka_status = "down"
        
        if location_res and location_res.status == 200 then
            location_status = "up"
        end
        
        if matching_res and matching_res.status == 200 then
            matching_status = "up"
        end

        if map_res and map_res.status == 200 then
            map_status = "up"
        end
        
        if kafka_res and kafka_res.status == 200 then
            kafka_status = "up"
        end
        
        local overall_status = "degraded"
        if location_status == "up" and matching_status == "up" and map_status == "up" then
            overall_status = "healthy"
        end
        
        -- Set HTTP status code based on overall health
        if overall_status ~= "healthy" then
            ngx.status = 503
        else
            ngx.status = 200
        end
        
        local timestamp = ngx.time()
        local formatted_time = ngx.http_time(timestamp)
        
        local response = string.format(
            '{"status":"%s","timestamp":"%s","services":{"kafka":"%s","location-service":"%s","matching-service":"%s","map-service":"%s","payment-service":"%s"}}',
            overall_status, formatted_time, location_status, location_status, matching_status, map_status,map_status
        )
        
        ngx.say(response)
    }
}

location = /ping {
    return 200 "pong\n";
    add_header Content-Type text/plain;
}