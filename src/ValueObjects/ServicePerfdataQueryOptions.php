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

namespace Statusengine\ValueObjects;


class ServicePerfdataQueryOptions extends QueryOptions {

    /**
     * @var string
     */
    private $hostname;

    /**
     * @var string
     */
    private $servicedescription;

    /**
     * @var string
     */
    private $metric;

    /**
     * @var int
     */
    private $points = 100;

    /**
     * @var int
     */
    private $start;

    /**
     * @var int
     */
    private $end = 0;

    /**
     * @var string
     */
    private $compression_algorithm = 'avg';

    /**
     * ServicePerfdataQueryOptions constructor.
     * @param $params
     */
    public function __construct($params) {
        parent::__construct($params);
        $this->params = $params;

        if (isset($this->params['hostname'])) {
            $this->hostname = $this->params['hostname'];
        }

        if (isset($this->params['servicedescription'])) {
            $this->servicedescription = $params['servicedescription'];
        }

        if (isset($this->params['metric'])) {
            $this->metric = $this->params['metric'];
        }

        if (isset($this->params['points'])) {
            $this->points = $this->params['points'];
        }

        $this->start = time();
        if (isset($this->params['start'])) {
            $this->start = (int)$this->params['start'];
        }

        if (isset($this->params['end'])) {
            $this->end = (int)$this->params['end'];
        }

        if (isset($this->params['compression_algorithm'])) {
            $this->compression_algorithm = $this->params['compression_algorithm'];
        }

        $this->fixCompressionAlgorithm();

    }

    /**
     * @return string
     */
    public function getHostname() {
        return $this->hostname;
    }

    /**
     * @return string
     */
    public function getServicedescription() {
        return $this->servicedescription;
    }

    /**
     * @return string
     */
    public function getMetric() {
        return $this->metric;
    }

    /**
     * @return int
     */
    public function getPoints() {
        return (int)$this->points;
    }

    /**
     * @return int
     */
    public function getStart() {
        return $this->start;
    }

    /**
     * @return int
     */
    public function getEnd() {
        return $this->end;
    }

    /**
     * @return bool
     */
    public function useAvg(){
        if($this->compression_algorithm == 'avg'){
            return true;
        }
        return false;
    }

    /**
     * @return bool
     */
    public function useMax(){
        if($this->compression_algorithm == 'max'){
            return true;
        }
        return false;
    }

    /**
     * @return bool
     */
    public function useMin(){
        if($this->compression_algorithm == 'min'){
            return true;
        }
        return false;
    }

    private function fixCompressionAlgorithm(){
        if(!in_array($this->compression_algorithm, ['avg', 'min', 'max'])){
            $this->compression_algorithm = 'avg';
        }
    }

}
