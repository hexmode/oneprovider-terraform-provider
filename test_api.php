<?php
/**
 * Make a call to OneProvider's API using stream context
 */
function call_api($api_key, $client_key, $http_method, $endpoint, $get = array(), $post = array()) {
    $base_url = 'https://api.oneprovider.com';

    if (!empty($get)) {
        $endpoint .= '?' . http_build_query($get);
    }

    $url = $base_url . $endpoint;

    $headers = [
        'Api-Key: ' . $api_key,
        'Client-Key: ' . $client_key,
        'X-Pretty-JSON: 1'
    ];

    $context = stream_context_create([
        'http' => [
            'method' => $http_method,
            'header' => $headers,
            'ignore_errors' => true
        ]
    ]);

    if ($http_method == 'POST' && !empty($post)) {
        $context = stream_context_create([
            'http' => [
                'method' => 'POST',
                'header' => array_merge($headers, ['Content-Type: application/x-www-form-urlencoded']),
                'content' => http_build_query($post),
                'ignore_errors' => true
            ]
        ]);
    }

    $result = file_get_contents($url, false, $context);
    return $result;
}

$api_key = getenv("ONEPROVIDER_API_KEY");
$client_key = getenv("ONEPROVIDER_CLIENT_KEY");

// Test SSH key creation
$new_key = call_api($api_key, $client_key, 'POST', '/vm/sshkey/new', array(), array(
    'key_name' => 'test-key',
    'key_value' => 'ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBBpwJwX/GzeS2OrZQRT13RRAKi5sfJ/ljpAj3rORt3PNHHWod9UvuKoTBDcnI1GmcTH+bpp8g2asHamBgbGW+Ug= mah@flynn'
));
echo "SSH Key create result:\n";
$json = json_decode($new_key);
var_dump($json);

$added = $json->response->key->uuid;
$new_key = call_api($api_key, $client_key, 'POST', '/vm/sshkey/delete', array(), array('ssh_key' => $added));
echo "Key removal result: $new_key\n";

// Test listing keys
$keys = call_api($api_key, $client_key, 'GET', '/vm/sshkeys/list');
echo "SSH Keys list:\n";
echo $keys . "\n";
