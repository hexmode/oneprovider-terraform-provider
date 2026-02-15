<?php
// Use curl instead of file_get_contents
function call_api_curl($api_key, $client_key, $http_method, $endpoint, $get = array(), $post = array()) {
    $base_url = 'https://api.oneprovider.com';
    
    if (!empty($get)) {
        $endpoint .= '?' . http_build_query($get);
    }

    $url = $base_url . $endpoint;
    
    $ch = curl_init();
    
    $headers = [
        'Api-Key: ' . $api_key,
        'Client-Key: ' . $client_key,
        'X-Pretty-JSON: 1'
    ];

    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
    curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);
    
    if ($http_method == 'POST' && !empty($post)) {
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($post));
    }

    $result = curl_exec($ch);
    curl_close($ch);
    
    return $result;
}

$api_key = getenv("ONEPROVIDER_API_KEY");
$client_key = getenv("ONEPROVIDER_CLIENT_KEY");

// Test with curl
$new_key = call_api_curl($api_key, $client_key, 'POST', '/vm/sshkey/new', array(), array(
    'key_name' => 'test-key-curl',
    'key_value' => 'ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBBpwJwX/GzeS2OrZQRT13RRAKi5sfJ/ljpAj3rORt3PNHHWod9UvuKoTBDcnI1GmcTH+bpp8g2asHamBgbGW+Ug= mah@flynn'
));
echo "curl result:\n";
echo $new_key . "\n";
?>
