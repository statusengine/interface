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
namespace Statusengine\Controller;


use Statusengine\Loader\ServiceLoaderInterface;
use Statusengine\ValueObjects\ServiceQueryOptions;

class Service extends Controller {

    /**
     * @var ServiceLoaderInterface
     */
    private $ServiceLoader;

    /**
     * Service constructor.
     * @param ServiceLoaderInterface $ServiceLoader
     */
    public function __construct(ServiceLoaderInterface $ServiceLoader) {
        $this->ServiceLoader = $ServiceLoader;
    }

    /**
     * @return array
     */
    public function index(ServiceQueryOptions $queryOptions) {
        return $this->ServiceLoader->getServiceList($queryOptions);
    }

    /**
     * @param ServiceQueryOptions $queryOptions
     * @return array
     */
    public function servicedetails(ServiceQueryOptions $queryOptions) {
        return $this->ServiceLoader->getServiceDetails($queryOptions);
    }

    /**
     * @return array
     */
    public function problems(ServiceQueryOptions $queryOptions) {
        return $this->ServiceLoader->getProblems($queryOptions);
    }

}