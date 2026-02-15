<?php
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, "https://api.oneprovider.com/vm/sshkey/new");
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query(["key_name" => "test", "key_value" => "ssh-rsa AAA test"]));
curl_setopt($ch, CURLOPT_HEADER, true);
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    "Api-Key: " . getenv("ONEPROVIDER_API_KEY"),
    "Client-Key: " . getenv("ONEPROVIDER_CLIENT_KEY"),
    "X-Pretty-JSON: 1",
    "Content-Type: application/x-www-form-urlencoded"
]);
$result = curl_exec($ch);
echo $result;
?>
