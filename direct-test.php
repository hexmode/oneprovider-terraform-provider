<?php
/**
  * Make a call to OneProvider's API
  */
function call_api($api_key, $client_key, $http_method, $endpoint, $get = array(), $post = array()) {
    if (!empty($get)) {
        $endpoint .= '?' . http_build_query($get);
    }

    $call = curl_init();
    curl_setopt($call, CURLOPT_URL, 'https://api.oneprovider.com' . $endpoint);
    curl_setopt($call, CURLOPT_HTTPHEADER, array(
        'Api-Key: ' . $api_key,
        'Client-Key: ' . $client_key,
        'X-Pretty-JSON: 1'
    ));
    curl_setopt($call, CURLOPT_RETURNTRANSFER, true);

    if ($http_method == 'POST') {
        curl_setopt($call, CURLOPT_POST, true);
        curl_setopt($call, CURLOPT_POSTFIELDS, http_build_query($post));
    } elseif ($http_method == 'DELETE') {
      curl_setopt($call, CURLOPT_CUSTOMREQUEST, $http_method);
    }

    $result = curl_exec($call);
    return $result;
}

$api_key = getenv("ONEPROVIDER_API_KEY");
$client_key = getenv("ONEPROVIDER_CLIENT_KEY");
$server_list = call_api($api_key, $client_key, 'POST', '/vm/sshkey/new', [], ["key_name" => "newkey", "key_value" => "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBBpwJwX/GzeS2OrZQRT13RRAKi5sfJ/ljpAj3rORt3PNHHWod9UvuKoTBDcnI1GmcTH+bpp8g2asHamBgbGW+Ug= mah@flynn"]);
echo $server_list;
