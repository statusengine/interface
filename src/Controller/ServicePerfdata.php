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

namespace Statusengine\Controller;

use Statusengine\Loader\ClusterLoaderInterface;
use Statusengine\Loader\Crate\ServicePerfdataLoader;
use Statusengine\Loader\ServicePerfdataLoaderInterfaceByConfig;
use Statusengine\ValueObjects\ServicePerfdataQueryOptions;

class ServicePerfdata extends Controller {

    /**
     * @var ServicePerfdataLoader
     */
    private $ServicePerfdataLoader;

    /**
     * ServicePerfdata constructor.
     * @param ServicePerfdataLoader|ServicePerfdataLoaderInterfaceByConfig $ServicePerfdataLoader
     */
    public function __construct($ServicePerfdataLoader) {
        $this->ServicePerfdataLoader = $ServicePerfdataLoader;
    }

    /**
     * @param ServicePerfdataQueryOptions $ServicePerfdataQueryOptions
     * @return array
     */
    public function index(ServicePerfdataQueryOptions $ServicePerfdataQueryOptions) {

        $perfdataResult = $this->ServicePerfdataLoader->getServicePerfdata($ServicePerfdataQueryOptions);

        $start = $ServicePerfdataQueryOptions->getStart();
        $end = $ServicePerfdataQueryOptions->getEnd();

        if (isset($perfdataResult[0]['timestamp'])) {
            $start = $perfdataResult[0]['timestamp'];

            $endKey = sizeof($perfdataResult) - 1;
            $end = $perfdataResult[$endKey]['timestamp'];
        }

        $result = [
            'perfdata' => $this->compress($ServicePerfdataQueryOptions, $perfdataResult),
            'start' => $start,
            'end' => $end
        ];
        return $result;
    }

    /**
     * @param ServicePerfdataQueryOptions $ServicePerfdataQueryOptions
     * @param $data
     * @return array
     */
    private function compress(ServicePerfdataQueryOptions $ServicePerfdataQueryOptions, $data) {
        if (sizeof($data) < $ServicePerfdataQueryOptions->getLimit()) {
            return $data;
        }

        $chunkSize = sizeof($data) / $ServicePerfdataQueryOptions->getLimit();
        $chunkSize = (int)$chunkSize;

        $result = [];
        foreach (array_chunk($data, $chunkSize) as $chunk) {
            $timestamps = [];
            $values = [];
            foreach ($chunk as $record) {
                $timestamps[] = $record['timestamp'];
                $values[] = $record['value'];
            }

            if($ServicePerfdataQueryOptions->useAvg()){
                $result[] = [
                    'timestamp' => array_sum($timestamps) / sizeof($timestamps),
                    'value' => array_sum($values) / sizeof($values)
                ];
            }

            if($ServicePerfdataQueryOptions->useMax()){
                $result[] = [
                    'timestamp' => array_sum($timestamps) / sizeof($timestamps),
                    'value' => max($values)
                ];
            }

            if($ServicePerfdataQueryOptions->useMin()){
                $result[] = [
                    'timestamp' => array_sum($timestamps) / sizeof($timestamps),
                    'value' => min($values)
                ];
            }

        }
        return $result;
    }


}