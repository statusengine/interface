<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler
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

namespace Statusengine\Loader\Elasticsearch;

use Elasticsearch\ClientBuilder;
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
        $Client = ClientBuilder::create()->setHosts($this->getHosts())->build();

        $params = [
            'index' => $this->Config->getElasticsearchIndex(),
            'type' => 'metric',
            'body' => [
                'size' => 10000,
                'sort' => [
                    'timestamp' => 'asc'
                ],
                'query' => [
                    'bool' => [
                        'must' => [
                            [
                                'term' => ['hostname' => $ServicePerfdataQueryOptions->getHostname()]
                            ],
                            [
                                'term' => ['service_description' => $ServicePerfdataQueryOptions->getServicedescription()]
                            ],
                            [
                                'term' => ['metric' => $ServicePerfdataQueryOptions->getMetric()]
                            ],
                            [
                                'range' => [
                                    'timestamp' => [
                                        'gt' => $ServicePerfdataQueryOptions->getEnd() * 1000, // >
                                        'lt' => $ServicePerfdataQueryOptions->getStart() * 1000 // <
                                    ]
                                ]
                            ]
                        ]
                    ]
                ]

            ]
        ];

        /*
        echo PHP_EOL;
        var_dump(sprintf('gt > %s', date('d-m-y H:i:s', $ServicePerfdataQueryOptions->getEnd())));
        var_dump(sprintf('lt < %s', date('d-m-y H:i:s', $ServicePerfdataQueryOptions->getStart())));

        echo PHP_EOL;
        print_r($params);*/

        $response = $Client->search($params);
        return $this->parseResponse($response);
    }


    /**
     * @return array
     */
    private function getHosts() {
        return [
            sprintf('%s:%s',
                $this->Config->getElasticsearchAddress(),
                $this->Config->getElasticsearchPort()
            )
        ];
    }

    /**
     * @param array $response
     * @return array
     */
    private function parseResponse($response){
        if(empty($response['hits']['hits'])){
            return [];
        }

        $result = [];
        foreach($response['hits']['hits'] as $record){
            $result[] = [
                'timestamp' => (int)$record['_source']['timestamp']/1000,
                'value' => $record['_source']['value'],
            ];
        }

        //Add 20 seconds to the last timestamp
        //so that the auto-refresh will not always
        //reload the same value from ES
        if(!empty($result)) {
            $endKey = sizeof($result) - 1;
            $result[$endKey]['timestamp'] += 20;
        }

        return $result;
    }

}
