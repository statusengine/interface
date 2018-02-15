<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Loader\Graphite;

use GuzzleHttp\Client;
use Statusengine\Config;
use Statusengine\Loader\ServicePerfdataLoaderInterfaceByConfig;
use Statusengine\ValueObjects\ServicePerfdataQueryOptions;

class ServicePerfdataLoader implements ServicePerfdataLoaderInterfaceByConfig {

    /**
     * @var Config
     */
    private $Config;

    public function __construct(Config $Config) {
        $this->Config = $Config;
    }


    /**
     * @param ServicePerfdataQueryOptions $ServicePerfdataQueryOptions
     * @return array
     */
    public function getServicePerfdata(ServicePerfdataQueryOptions $ServicePerfdataQueryOptions) {

        $client = new Client(['base_uri' => $this->Config->getGraphiteUrl()]);

        $start = time() - $ServicePerfdataQueryOptions->getEnd();
        if($start < 0){
            $start = 3600 * 2.5;
        }

        $options = [
            'allow_redirects' => true,
            'stream_context' => [
                'ssl' => [
                    'allow_self_signed' => $this->Config->getGraphiteAllowSelfSignedCertificates()
                ]
            ],
            'query' => [
                'target' => sprintf(
                    '%s.%s.%s.%s',
                    $this->Config->getGraphitePrefix(),
                    $this->replaceIllegalCharacters($ServicePerfdataQueryOptions->getHostname()),
                    $this->replaceIllegalCharacters($ServicePerfdataQueryOptions->getServicedescription()),
                    $this->replaceIllegalCharacters($ServicePerfdataQueryOptions->getMetric())
                    ),
                'from'  => sprintf('-%ss', $start),
                //'from'  => date('H:i_Ymd', $ServicePerfdataQueryOptions->getEnd()),
                //'until' => sprintf('-%ss', time() - $ServicePerfdataQueryOptions->getStart()),
                //'until' => date('H:i_Ymd', $ServicePerfdataQueryOptions->getStart()),
                'noNullPoints' => 'true',
                'format' => 'json'
            ]
        ];

        if($this->Config->getGraphiteUseBasicAuth()){
            $options['auth'] = [
                $this->Config->getGraphiteUser(),
                $this->Config->getGraphitePassword()
            ];
        }

        $response = $client->request('GET', '/render', $options);
        $statusCode = $response->getStatusCode();
        if($statusCode == 200){
            $body = $response->getBody();
            $content = $body->getContents();
            $content = json_decode($content);

            if(isset($content[0]) && property_exists($content[0], 'datapoints')){
                $result = $this->reformAndFilterNull($content[0]->datapoints);
                unset($content);

                return $result;
            }

        }


        return [];
    }

    public function reformAndFilterNull($data){
        //For older Graphite version, we need to implement noNullPoints
        $result = [];
        foreach($data as $key => $record){
            if($record[0] !== null){
                $result[] = [
                    'timestamp' => $record[1],
                    'value'     => $record[0],
                ];
            }
            unset($data[$key]);
        }

        //Add 20 seconds to the last timestamp
        //so that the auto-refresh will not alway
        //reload the same value from graphite
        if(!empty($result)) {
            $endKey = sizeof($result) - 1;
            $result[$endKey]['timestamp'] += 20;
        }

        return $result;
    }

    public function replaceIllegalCharacters($str){
        $regex = $this->Config->getGraphiteIllegalCharacters();
        return preg_replace($regex, '_', $str);
    }

}
