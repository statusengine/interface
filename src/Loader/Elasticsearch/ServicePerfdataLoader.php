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

namespace Statusengine\Loader\Elasticsearch;

use Elasticsearch\ClientBuilder;
use Statusengine\Backend\Elasticsearch\Pattern;
use Statusengine\Config;
use Statusengine\Loader\ServicePerfdataLoaderInterfaceByConfig;
use Statusengine\ValueObjects\ServicePerfdataQueryOptions;

class ServicePerfdataLoader implements ServicePerfdataLoaderInterfaceByConfig {

    /**
     * @var Config
     */
    private $Config;

    /**
     * @var string
     * Keeping the search context alive
     */
    private $scroll = '15s';

    /**
     * @var string
     */
    private $index;

    /**
     * @var int
     * Number of results per shard
     */
    private $size = 500;

    public function __construct(Config $Config) {
        $this->Config = $Config;
        $this->index = $this->Config->getElasticsearchIndex();
    }


    /**
     * @param ServicePerfdataQueryOptions $ServicePerfdataQueryOptions
     * @return array
     */
    public function getServicePerfdata(ServicePerfdataQueryOptions $ServicePerfdataQueryOptions) {
        $Client = ClientBuilder::create()->setHosts($this->getHosts())->build();

        $Pattern = new Pattern($this->Config);
        $result = [];
        foreach ($Pattern->getPossibleIndices($ServicePerfdataQueryOptions) as $index) {
            $indexExists = $Client->indices()->exists(['index' => $index]);
            if ($indexExists) {
                $params = $this->getParams($ServicePerfdataQueryOptions, $index);
                $response = $Client->search($params);
                while (isset($response['hits']['hits']) && sizeof($response['hits']['hits']) > 0) {
                    foreach ($response['hits']['hits'] as $record) {
                        $result[] = [
                            'timestamp' => (int)$record['_source']['@timestamp'] / 1000,
                            'value' => $record['_source']['value'],
                        ];
                    }

                    //Fetch rest of data - if any
                    $response = $Client->scroll([
                            "scroll_id" => $response['_scroll_id'],
                            "scroll" => $this->scroll
                        ]
                    );
                }
                unset($response);
            }
        }
        return $result;
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
     * @param ServicePerfdataQueryOptions $ServicePerfdataQueryOptions
     * @param string $index
     * @return array
     */
    private function getParams(ServicePerfdataQueryOptions $ServicePerfdataQueryOptions, $index) {
        $params = [
            'scroll' => $this->scroll,
            'size' => $this->size,
            'index' => $index,
            'type' => 'metric',
            'body' => [
                'sort' => [
                    '@timestamp' => 'asc'
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
                                    '@timestamp' => [
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
        return $params;
    }

}
